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

package resource

import (
	"github.com/deepflowio/deepflow/server/controller/db/mysql"
	mysqlmodel "github.com/deepflowio/deepflow/server/controller/db/mysql/model"
)

func GetVPCs(orgID int, filter map[string]interface{}) ([]*mysqlmodel.VPC, error) {
	dbInfo, err := mysql.GetDB(orgID)
	if err != nil {
		return nil, err
	}
	db := dbInfo.DB
	if _, ok := filter["name"]; ok {
		db = db.Where("name = ?", filter["name"])
	}
	var vpcs []*mysqlmodel.VPC
	if err := db.Where("deleted_at IS NULL").Order("created_at DESC").Find(&vpcs).Error; err != nil {
		return nil, err
	}
	return vpcs, nil
}
