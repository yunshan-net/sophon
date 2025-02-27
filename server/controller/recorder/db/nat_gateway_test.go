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

package db

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	metadbmodel "github.com/deepflowio/deepflow/server/controller/db/metadb/model"
)

func newDBNATGateway() *metadbmodel.NATGateway {
	return &metadbmodel.NATGateway{Base: metadbmodel.Base{Lcuuid: uuid.New().String()}, Name: uuid.New().String()}
}

func (t *SuiteTest) TestAddNATGatewayBatchSuccess() {
	operator := NewNATGateway()
	itemToAdd := newDBNATGateway()

	_, ok := operator.AddBatch([]*metadbmodel.NATGateway{itemToAdd})
	assert.True(t.T(), ok)

	var addedItem *metadbmodel.NATGateway
	t.db.Where("lcuuid = ?", itemToAdd.Lcuuid).Find(&addedItem)
	assert.Equal(t.T(), addedItem.Name, itemToAdd.Name)

	t.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&metadbmodel.NATGateway{})
}

func (t *SuiteTest) TestUpdateNATGatewaySuccess() {
	operator := NewNATGateway()
	addedItem := newDBNATGateway()
	result := t.db.Create(&addedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(1))

	updateInfo := map[string]interface{}{"name": uuid.New().String()}
	_, ok := operator.Update(addedItem.Lcuuid, updateInfo)
	assert.True(t.T(), ok)

	var updatedItem *metadbmodel.NATGateway
	t.db.Where("lcuuid = ?", addedItem.Lcuuid).Find(&updatedItem)
	assert.Equal(t.T(), updatedItem.Name, updateInfo["name"])

	t.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&metadbmodel.NATGateway{})
}

func (t *SuiteTest) TestDeleteNATGatewayBatchSuccess() {
	operator := NewNATGateway()
	addedItem := newDBNATGateway()
	result := t.db.Create(&addedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(1))

	assert.True(t.T(), operator.DeleteBatch([]string{addedItem.Lcuuid}))
	var deletedItem *metadbmodel.NATGateway
	result = t.db.Where("lcuuid = ?", addedItem.Lcuuid).Find(&deletedItem)
	assert.Equal(t.T(), result.RowsAffected, int64(0))
}

func (t *SuiteTest) TestNATGatewayCreateAndFind() {
	lcuuid := uuid.New().String()
	natGateway := &metadbmodel.NATGateway{
		Base: metadbmodel.Base{Lcuuid: lcuuid},
	}
	t.db.Create(natGateway)
	var resultNATGateway *metadbmodel.NATGateway
	err := t.db.Where("lcuuid = ? and name='' and label='' and floating_ips='' "+
		"and az='' and region='' and uid=''", lcuuid).First(&resultNATGateway).Error
	assert.Equal(t.T(), nil, err)
	assert.Equal(t.T(), natGateway.Base.Lcuuid, resultNATGateway.Base.Lcuuid)

	resultNATGateway = new(metadbmodel.NATGateway)
	t.db.Where("lcuuid = ?", lcuuid).Find(&resultNATGateway)
	assert.Equal(t.T(), natGateway.Base.Lcuuid, resultNATGateway.Base.Lcuuid)

	resultNATGateway = new(metadbmodel.NATGateway)
	result := t.db.Where("lcuuid = ? and name = null", lcuuid).Find(&resultNATGateway)
	assert.Equal(t.T(), nil, result.Error)
	assert.Equal(t.T(), int64(0), result.RowsAffected)
}
