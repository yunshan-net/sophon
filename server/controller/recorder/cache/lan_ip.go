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

package cache

import (
	cloudmodel "github.com/deepflowio/deepflow/server/controller/cloud/model"
	"github.com/deepflowio/deepflow/server/controller/db/mysql"
	. "github.com/deepflowio/deepflow/server/controller/recorder/common"
)

func (b *DiffBaseDataSet) addLANIP(dbItem *mysql.LANIP, seq int, toolDataSet *ToolDataSet) {
	subnetLcuuid, _ := toolDataSet.GetSubnetLcuuidByID(dbItem.SubnetID)
	b.LANIPs[dbItem.Lcuuid] = &LANIP{
		DiffBase: DiffBase{
			Sequence: seq,
			Lcuuid:   dbItem.Lcuuid,
		},
		SubDomainLcuuid: dbItem.SubDomain,
		SubnetLcuuid:    subnetLcuuid,
	}
	b.GetLogFunc()(addDiffBase(RESOURCE_TYPE_LAN_IP_EN, b.LANIPs[dbItem.Lcuuid]))
}

func (b *DiffBaseDataSet) deleteLANIP(lcuuid string) {
	delete(b.LANIPs, lcuuid)
	log.Info(deleteDiffBase(RESOURCE_TYPE_LAN_IP_EN, lcuuid))
}

type LANIP struct {
	DiffBase
	SubDomainLcuuid string `json:"sub_domain_lcuuid"`
	SubnetLcuuid    string `json:"subnet_lcuuid"`
}

func (l *LANIP) Update(cloudItem *cloudmodel.IP) {
	l.SubDomainLcuuid = cloudItem.SubDomainLcuuid
	l.SubnetLcuuid = cloudItem.SubnetLcuuid
	log.Info(updateDiffBase(RESOURCE_TYPE_LAN_IP_EN, l))
}
