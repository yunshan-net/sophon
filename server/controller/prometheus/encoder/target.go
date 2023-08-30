/**
 * Copyright (c) 2023 Yunshan Networks
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package encoder

import (
	"sync"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/golang/protobuf/proto"

	"github.com/deepflowio/deepflow/message/controller"
	"github.com/deepflowio/deepflow/server/controller/common"
	"github.com/deepflowio/deepflow/server/controller/db/mysql"
	"github.com/deepflowio/deepflow/server/controller/prometheus/cache"
)

type targetKeyToID map[cache.TargetKey]int

// 缓存资源可用于分配的ID，提供ID的刷新、分配、回收接口
type target struct {
	lock         sync.Mutex
	resourceType string
	keyToID      sync.Map
	descIDAllocator
}

func newTarget(max int) *target {
	ln := &target{
		resourceType: "target",
	}
	// both recorder and prometheus need to insert data into prometheus_target, they equally share the id pool of prometheus_target.
	// recorder uses ids [1, max/2+max%2], prometheus uses ids [max/2+max%2+1, max].
	ln.descIDAllocator = newDescIDAllocator(ln.resourceType, max/2+max%2+1, max)
	ln.rawDataProvider = ln
	return ln
}

func (ln *target) getID(key cache.TargetKey) (int, bool) {
	if item, ok := ln.keyToID.Load(key); ok {
		return item.(int), true
	}
	return 0, false
}

func (ln *target) refresh(args ...interface{}) error {
	ln.lock.Lock()
	defer ln.lock.Unlock()

	return ln.descIDAllocator.refresh()
}

func (ln *target) encode(ts []*controller.PrometheusTargetRequest) ([]*controller.PrometheusTarget, error) {
	ln.lock.Lock()
	defer ln.lock.Unlock()

	resp := make([]*controller.PrometheusTarget, 0)
	var dbToAdd []*mysql.PrometheusTarget
	for i := range ts {
		t := ts[i]
		di, err := getDomainInfoByVTapID(int(t.GetVtapId()))
		if err != nil {
			return nil, err
		}
		ins := t.GetInstance()
		job := t.GetJob()
		if id, ok := ln.keyToID.Load(cache.NewTargetKey(ins, job)); ok {
			resp = append(resp, &controller.PrometheusTarget{
				Id:       proto.Uint32(uint32(id.(int))),
				Instance: &ins,
				Job:      &job,
			})
			continue
		}
		dbToAdd = append(dbToAdd, &mysql.PrometheusTarget{ // TODO  id 复用
			Base:         mysql.Base{Lcuuid: common.GenerateUUID(ins + job + "prometheus")},
			CreateMethod: common.PROMETHEUS_TARGET_CREATE_METHOD_PROMETHEUS,
			Instance:     ins,
			Job:          job,
			Domain:       di.domain,
			SubDomain:    di.subDomain,
		})
	}
	if len(dbToAdd) == 0 {
		return resp, nil
	}
	ids, err := ln.allocate(len(dbToAdd))
	if err != nil {
		return nil, err
	}
	for i := range ids {
		dbToAdd[i].ID = ids[i]
	}
	err = addBatch(dbToAdd, ln.resourceType)
	if err != nil {
		log.Errorf("add %s error: %s", ln.resourceType, err.Error())
		return nil, err
	}
	for i := range dbToAdd {
		id := dbToAdd[i].ID
		k := cache.NewTargetKey(dbToAdd[i].Instance, dbToAdd[i].Job)
		ln.keyToID.Store(k, id)
		resp = append(resp, &controller.PrometheusTarget{
			Id:       proto.Uint32(uint32(id)),
			Instance: &k.Instance,
			Job:      &k.Job,
		})
	}
	return resp, nil
}

func (ln *target) load() (ids mapset.Set[int], err error) {
	var items []*mysql.PrometheusTarget
	err = mysql.Db.Where(&mysql.PrometheusTarget{CreateMethod: common.PROMETHEUS_TARGET_CREATE_METHOD_PROMETHEUS}).Find(&items).Error
	if err != nil {
		log.Errorf("db query %s failed: %v", ln.resourceType, err)
		return nil, err
	}
	inUseIDSet := mapset.NewSet[int]()
	for _, item := range items {
		inUseIDSet.Add(item.ID)
		ln.keyToID.Store(cache.NewTargetKey(item.Instance, item.Job), item.ID)
	}
	return inUseIDSet, nil
}

func (ln *target) check(ids []int) (inUseIDs []int, err error) {
	var dbItems []*mysql.PrometheusTarget
	err = mysql.Db.Unscoped().Where("id IN ?", ids).Find(&dbItems).Error
	if err != nil {
		log.Errorf("db query %s failed: %v", ln.resourceType, err)
		return
	}
	if len(dbItems) != 0 {
		for _, item := range dbItems {
			inUseIDs = append(inUseIDs, item.ID)
		}
		log.Infof("%s ids: %+v are in use.", ln.resourceType, inUseIDs)
	}
	return
}

type domainInfo struct {
	domain    string
	subDomain string
}

func getDomainInfoByVTapID(id int) (*domainInfo, error) {
	vtap, err := getByID[mysql.VTap](id)
	if err != nil {
		return nil, err
	}
	di := new(domainInfo)
	switch common.VTAP_TYPE_TO_DEVICE_TYPE[vtap.Type] {
	case common.VIF_DEVICE_TYPE_HOST:
		var device *mysql.Host
		device, err = getByID[mysql.Host](vtap.LaunchServerID)
		di.domain = device.Domain
	case common.VIF_DEVICE_TYPE_VM:
		var device *mysql.VM
		device, err = getByID[mysql.VM](vtap.LaunchServerID)
		di.domain = device.Domain
	case common.VIF_DEVICE_TYPE_POD_NODE:
		var device *mysql.PodNode
		device, err = getByID[mysql.PodNode](vtap.LaunchServerID)
		di.domain = device.Domain
		di.subDomain = device.SubDomain
	case common.VIF_DEVICE_TYPE_POD:
		var device *mysql.Pod
		device, err = getByID[mysql.Pod](vtap.LaunchServerID)
		di.domain = device.Domain
		di.subDomain = device.SubDomain
	}
	return di, err
}

func getByID[T any](id int) (*T, error) {
	var item *T
	err := mysql.Db.Where("id = ?", id).Find(&item).Error
	return item, err
}
