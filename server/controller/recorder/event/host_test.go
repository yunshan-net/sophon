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

package event

import (
	"testing"

	metadbmodel "github.com/deepflowio/deepflow/server/controller/db/metadb/model"
	"github.com/deepflowio/deepflow/server/controller/recorder/cache"
	"github.com/deepflowio/deepflow/server/controller/recorder/cache/diffbase"
	"github.com/deepflowio/deepflow/server/controller/recorder/cache/tool"
	"github.com/deepflowio/deepflow/server/libs/eventapi"
	"github.com/stretchr/testify/assert"
)

func TestHost_ProduceByAdd(t *testing.T) {
	dataSet := tool.NewDataSet()
	type args struct {
		items []*metadbmodel.Host
	}
	tests := []struct {
		name     string
		h        *Host
		args     args
		wantID   uint32
		wantName string
	}{
		{
			name: "add success",
			h:    NewHost(dataSet, NewEventQueue()),
			args: args{
				items: []*metadbmodel.Host{
					{
						Base: metadbmodel.Base{ID: 1},
						Name: "host",
					},
				},
			},
			wantID:   1,
			wantName: "host",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.h.ProduceByAdd(tt.args.items)

			e := tt.h.ManagerComponent.Queue.Get().(*eventapi.ResourceEvent)
			assert.Equal(t, tt.wantID, e.InstanceID)
			assert.Equal(t, tt.wantName, e.InstanceName)
		})
	}
}

func TestHost_ProduceByDelete(t *testing.T) {
	type args struct {
		lcuuids []string
	}
	tests := []struct {
		name     string
		cache    *cache.Cache
		h        *Host
		args     args
		prepare  func(*cache.Cache)
		wantID   uint32
		wantName string
	}{
		{
			name: "delete success",
			cache: &cache.Cache{
				DiffBaseDataSet: diffbase.NewDataSet(),
				ToolDataSet:     tool.NewDataSet(),
			},
			args: args{
				lcuuids: []string{"host_lcuuid"},
			},
			prepare: func(cache *cache.Cache) {
				cache.AddRegion(&metadbmodel.Region{Base: metadbmodel.Base{ID: 1, Lcuuid: "region_lcuuid"}})
				cache.AddAZ(&metadbmodel.AZ{Base: metadbmodel.Base{ID: 2, Lcuuid: "az_lcuuid"}})
				cache.AddHost(&metadbmodel.Host{
					Base: metadbmodel.Base{
						ID:     1,
						Lcuuid: "host_lcuuid",
					},
					Name:   "host",
					Region: "region_lcuuid",
					AZ:     "az_lcuuid",
				})
			},
			wantID:   1,
			wantName: "host",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare(tt.cache)
			tt.h = NewHost(tt.cache.ToolDataSet, NewEventQueue())
			tt.h.ProduceByDelete(tt.args.lcuuids)

			e := tt.h.ManagerComponent.Queue.Get().(*eventapi.ResourceEvent)
			assert.Equal(t, tt.wantID, e.InstanceID)
			assert.Equal(t, tt.wantName, e.InstanceName)
		})
	}
}
