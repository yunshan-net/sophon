/*
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

package listener

import (
	cloudmodel "github.com/deepflowio/deepflow/server/controller/cloud/model"
	"github.com/deepflowio/deepflow/server/controller/db/mysql"
	"github.com/deepflowio/deepflow/server/controller/recorder/cache"
)

type VPC struct {
	cache *cache.Cache
}

func NewVPC(c *cache.Cache) *VPC {
	return &VPC{
		cache: c,
	}
}

func (v *VPC) OnUpdaterAdded(addedDBItems []*mysql.VPC) {
	v.cache.AddVPCs(addedDBItems)
}

func (v *VPC) OnUpdaterUpdated(cloudItem *cloudmodel.VPC, diffBase *cache.VPC) {
	diffBase.Update(cloudItem)
}

func (v *VPC) OnUpdaterDeleted(lcuuids []string) {
	v.cache.DeleteVPCs(lcuuids)
}
