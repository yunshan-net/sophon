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

package tagrecorder

import (
	"github.com/deepflowio/deepflow/server/controller/db/metadb"
	metadbmodel "github.com/deepflowio/deepflow/server/controller/db/metadb/model"
)

type ChPrometheusMetricName struct {
	UpdaterComponent[metadbmodel.ChPrometheusMetricName, IDKey]
}

func NewChPrometheusMetricNames() *ChPrometheusMetricName {
	updater := &ChPrometheusMetricName{
		newUpdaterComponent[metadbmodel.ChPrometheusMetricName, IDKey](
			RESOURCE_TYPE_CH_METRIC_NAME,
		),
	}

	updater.updaterDG = updater
	return updater
}

func (l *ChPrometheusMetricName) generateNewData(db *metadb.DB) (map[IDKey]metadbmodel.ChPrometheusMetricName, bool) {
	var prometheusMetricName []metadbmodel.PrometheusMetricName
	err := db.Unscoped().Find(&prometheusMetricName).Error
	if err != nil {
		log.Errorf(dbQueryResourceFailed(l.resourceTypeName, err), db.LogPrefixORGID)
		return nil, false
	}

	keyToItem := make(map[IDKey]metadbmodel.ChPrometheusMetricName)
	for _, metricName := range prometheusMetricName {
		keyToItem[IDKey{ID: metricName.ID}] = metadbmodel.ChPrometheusMetricName{
			ID:   metricName.ID,
			Name: metricName.Name,
		}
	}
	return keyToItem, true
}

func (l *ChPrometheusMetricName) generateKey(dbItem metadbmodel.ChPrometheusMetricName) IDKey {
	return IDKey{ID: dbItem.ID}
}

func (l *ChPrometheusMetricName) generateUpdateInfo(oldItem, newItem metadbmodel.ChPrometheusMetricName) (map[string]interface{}, bool) {
	updateInfo := make(map[string]interface{})
	if oldItem.Name != newItem.Name {
		updateInfo["name"] = newItem.Name
	}
	if len(updateInfo) > 0 {
		return updateInfo, true
	}
	return nil, false
}
