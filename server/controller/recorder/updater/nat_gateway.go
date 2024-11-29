/*
 * Copyright (c) 2024 Yunshan Networks
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

package updater

import (
	cloudmodel "github.com/deepflowio/deepflow/server/controller/cloud/model"
	ctrlrcommon "github.com/deepflowio/deepflow/server/controller/common"
	mysqlmodel "github.com/deepflowio/deepflow/server/controller/db/mysql/model"
	"github.com/deepflowio/deepflow/server/controller/recorder/cache"
	"github.com/deepflowio/deepflow/server/controller/recorder/cache/diffbase"
	"github.com/deepflowio/deepflow/server/controller/recorder/db"
	"github.com/deepflowio/deepflow/server/controller/recorder/pubsub/message"
)

type NATGateway struct {
	UpdaterBase[
		cloudmodel.NATGateway,
		*diffbase.NATGateway,
		*mysqlmodel.NATGateway,
		mysqlmodel.NATGateway,
		*message.NATGatewayAdd,
		message.NATGatewayAdd,
		*message.NATGatewayUpdate,
		message.NATGatewayUpdate,
		*message.NATGatewayFieldsUpdate,
		message.NATGatewayFieldsUpdate,
		*message.NATGatewayDelete,
		message.NATGatewayDelete]
}

func NewNATGateway(wholeCache *cache.Cache, cloudData []cloudmodel.NATGateway) *NATGateway {
	updater := &NATGateway{
		newUpdaterBase[
			cloudmodel.NATGateway,
			*diffbase.NATGateway,
			*mysqlmodel.NATGateway,
			mysqlmodel.NATGateway,
			*message.NATGatewayAdd,
			message.NATGatewayAdd,
			*message.NATGatewayUpdate,
			message.NATGatewayUpdate,
			*message.NATGatewayFieldsUpdate,
			message.NATGatewayFieldsUpdate,
			*message.NATGatewayDelete,
		](
			ctrlrcommon.RESOURCE_TYPE_NAT_GATEWAY_EN,
			wholeCache,
			db.NewNATGateway(),
			wholeCache.DiffBaseDataSet.NATGateways,
			cloudData,
		),
	}
	updater.dataGenerator = updater
	updater.initDBOperator()
	return updater
}

func (g *NATGateway) getDiffBaseByCloudItem(cloudItem *cloudmodel.NATGateway) (diffBase *diffbase.NATGateway, exists bool) {
	diffBase, exists = g.diffBaseData[cloudItem.Lcuuid]
	return
}

func (g *NATGateway) generateDBItemToAdd(cloudItem *cloudmodel.NATGateway) (*mysqlmodel.NATGateway, bool) {
	vpcID, exists := g.cache.ToolDataSet.GetVPCIDByLcuuid(cloudItem.VPCLcuuid)
	if !exists {
		log.Error(resourceAForResourceBNotFound(
			ctrlrcommon.RESOURCE_TYPE_VPC_EN, cloudItem.VPCLcuuid,
			ctrlrcommon.RESOURCE_TYPE_NAT_GATEWAY_EN, cloudItem.Lcuuid,
		), g.metadata.LogPrefixes)
		return nil, false
	}

	dbItem := &mysqlmodel.NATGateway{
		Name:        cloudItem.Name,
		Label:       cloudItem.Label,
		UID:         cloudItem.Label,
		FloatingIPs: cloudItem.FloatingIPs,
		Domain:      g.metadata.Domain.Lcuuid,
		Region:      cloudItem.RegionLcuuid,
		VPCID:       vpcID,
	}
	dbItem.Lcuuid = cloudItem.Lcuuid
	return dbItem, true
}

func (g *NATGateway) getUpdateableFields() []string {
	return []string{"name", "region", "floating_ips"}
}

func (g *NATGateway) generateUpdateInfo(diffBase *diffbase.NATGateway, cloudItem *cloudmodel.NATGateway) (*message.NATGatewayFieldsUpdate, map[string]interface{}, bool) {
	structInfo := new(message.NATGatewayFieldsUpdate)
	mapInfo := make(map[string]interface{})
	if diffBase.Name != cloudItem.Name {
		mapInfo["name"] = cloudItem.Name
		structInfo.Name.Set(diffBase.Name, cloudItem.Name)
	}
	if diffBase.RegionLcuuid != cloudItem.RegionLcuuid {
		mapInfo["region"] = cloudItem.RegionLcuuid
		structInfo.RegionLcuuid.Set(diffBase.RegionLcuuid, cloudItem.RegionLcuuid)
	}
	if diffBase.FloatingIPs != cloudItem.FloatingIPs {
		mapInfo["floating_ips"] = cloudItem.FloatingIPs
		structInfo.FloatingIPs.Set(diffBase.FloatingIPs, cloudItem.FloatingIPs)
	}

	return structInfo, mapInfo, len(mapInfo) > 0
}

func (g *NATGateway) setUpdatedFields(dbItem *mysqlmodel.NATGateway, updateInfo *message.NATGatewayFieldsUpdate) {
	if updateInfo.Name.IsDifferent() {
		dbItem.Name = updateInfo.Name.GetNew()
	}
	if updateInfo.RegionLcuuid.IsDifferent() {
		dbItem.Region = updateInfo.RegionLcuuid.GetNew()
	}
	if updateInfo.FloatingIPs.IsDifferent() {
		dbItem.FloatingIPs = updateInfo.FloatingIPs.GetNew()
	}
}
