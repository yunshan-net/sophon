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

package service

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/deepflowio/deepflow/server/controller/db/metadb"
	metadbmodel "github.com/deepflowio/deepflow/server/controller/db/metadb/model"
	httpcommon "github.com/deepflowio/deepflow/server/controller/http/common"
	"github.com/deepflowio/deepflow/server/controller/http/common/response"
	"github.com/deepflowio/deepflow/server/controller/model"
)

var namespace = uuid.NewSHA1(uuid.NameSpaceDNS, []byte("native"))

func GetNativeFields(orgID int) (resp []metadbmodel.NativeField, err error) {
	dbInfo, err := metadb.GetDB(orgID)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	db := dbInfo.DB
	if err := db.Find(&resp).Error; err != nil {
		log.Error(err)
		return nil, err
	}
	return
}

func CreateNativeField(orgID int, nativeFieldCreate model.NativeFieldCreate) ([]metadbmodel.NativeField, error) {
	database := nativeFieldCreate.Database
	tableName := nativeFieldCreate.TableName
	name := nativeFieldCreate.Name
	lcuuid := uuid.NewSHA1(namespace, []byte(database+tableName+name))
	lcuuidStr := lcuuid.String()
	oldNativeFields, err := GetNativeFields(orgID)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	for _, nativeField := range oldNativeFields {
		if lcuuidStr == nativeField.Lcuuid {
			return nil, response.ServiceError(
				httpcommon.RESOURCE_ALREADY_EXIST, fmt.Sprintf("field (%s) is already existed", lcuuidStr),
			)
		}
	}
	// call zhuofeng

	dbInfo, err := metadb.GetDB(orgID)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	db := dbInfo.DB
	nativeField := metadbmodel.NativeField{}
	nativeField.Database = database
	nativeField.TableName = tableName
	nativeField.Name = name
	nativeField.FieldName = nativeFieldCreate.FieldName
	nativeField.FieldValueType = nativeFieldCreate.FieldValueType
	nativeField.TeamID = nativeFieldCreate.TeamID
	nativeField.Lcuuid = lcuuidStr
	if err := db.Create(&nativeField).Error; err != nil {
		return nil, err
	}
	newNativeFields, err := GetNativeFields(orgID)
	if err != nil {
		return nil, err
	}
	return newNativeFields, nil
}

func DeleteNativeField(orgID int, lcuuid string) ([]metadbmodel.NativeField, error) {
	dbInfo, err := metadb.GetDB(orgID)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	db := dbInfo.DB
	var nativeField metadbmodel.NativeField

	if ret := db.Where("lcuuid = ?", lcuuid).First(&nativeField); ret.Error != nil {
		return nil, response.ServiceError(
			httpcommon.RESOURCE_NOT_FOUND, fmt.Sprintf("field (%s) not found", lcuuid),
		)
	}
	// call zhuofeng

	if err := db.Delete(&nativeField).Error; err != nil {
		log.Error(err)
		return nil, err
	}
	newNativeFields, err := GetNativeFields(orgID)
	if err != nil {
		return nil, err
	}
	return newNativeFields, err
}
