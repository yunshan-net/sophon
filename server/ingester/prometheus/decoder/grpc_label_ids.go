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

package decoder

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/cornelk/hashmap"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"

	"github.com/deepflowio/deepflow/message/trident"
	"github.com/deepflowio/deepflow/server/ingester/common"
	"github.com/deepflowio/deepflow/server/ingester/ingesterctl"
	"github.com/deepflowio/deepflow/server/libs/ckdb"
	"github.com/deepflowio/deepflow/server/libs/debug"
	"github.com/deepflowio/deepflow/server/libs/grpc"
	"github.com/deepflowio/deepflow/server/libs/utils"
)

const (
	METRICID_OFFSET = 32 // when generate columnIndexKey/metricTargetPairKey, high32 is metricID, low32 can be labelNameID/targetID

	MAX_ORG_COUNT = ckdb.MAX_ORG_ID + 1
)

func uint64ToFloat64(i uint64) float64 {
	return *(*float64)(unsafe.Pointer(&i))
}

func float64ToUint64(f float64) uint64 {
	return *(*uint64)(unsafe.Pointer(&f))
}

func columnIndexKey(metricID, labelNameID uint32) uint64 {
	return uint64(metricID)<<METRICID_OFFSET | uint64(labelNameID)
}

func nameValueKey(nameID, valueID uint32) uint64 {
	return uint64(nameID)<<32 | uint64(valueID)
}

func (t *PrometheusLabelTable) QueryMetricID(orgId uint16, metricName string) (uint32, bool) {
	if value, exists := t.metricNameIDs[orgId].Get(metricName); exists {
		return t.getId(value)
	}
	return 0, false
}

func (t *PrometheusLabelTable) QueryLabelNameID(orgId uint16, labelName string) (uint32, bool) {
	if value, exists := t.labelNameIDs[orgId].Get(labelName); exists {
		return t.getId(value)
	}
	return 0, false
}

func (t *PrometheusLabelTable) QueryLabelValueID(orgId uint16, labelValue string) (uint32, bool) {
	if value, exists := t.labelValueIDs[orgId].Get(labelValue); exists {
		return t.getId(value)
	}
	return 0, false
}

func (t *PrometheusLabelTable) QueryLabelNameValue(orgId uint16, nameId, valueId uint32) bool {
	_, exists := t.labelNameValues[orgId].Get(nameValueKey(nameId, valueId))
	return exists
}

func (t *PrometheusLabelTable) QueryColumnIndex(orgId uint16, metricID, labelNameID uint32) (uint32, bool) {
	return t.labelColumnIndexs[orgId].Get(columnIndexKey(metricID, labelNameID))
}

type RequestCounter struct {
	RequestCount        int64  `statsd:"request-count"`
	RequestTotalDelayNs int64  `statsd:"request-total-delay-ns"`
	ResponseFailed      int64  `statsd:"response-failed"`
	RequestLabelsCount  int64  `statsd:"request-labels-count"`
	ResponseLabelsCount int64  `statsd:"response-labels-count"`
	MetricUnknown       uint64 `statsd:"metric-unknown"`
	LabelNameUnknown    uint64 `statsd:"label-name-unknown"`
	LabelValueUnknown   uint64 `statsd:"label-value-unknown"`
	CacheExpiration     uint64 `statsd:"cache-expiration-count"`
}

func (t *PrometheusLabelTable) GetCounter() interface{} {
	var counter *RequestCounter
	counter, t.counter = t.counter, &RequestCounter{}
	return counter
}

type PrometheusLabelTable struct {
	ctlIP       string
	GrpcSession *grpc.GrpcSession

	metricNameIDs     [MAX_ORG_COUNT]*hashmap.Map[string, uint64]
	labelNameIDs      [MAX_ORG_COUNT]*hashmap.Map[string, uint64]
	labelValueIDs     [MAX_ORG_COUNT]*hashmap.Map[string, uint64]
	labelNameValues   [MAX_ORG_COUNT]*hashmap.Map[uint64, struct{}]
	labelColumnIndexs [MAX_ORG_COUNT]*hashmap.Map[uint64, uint32]
	cacheExpiration   int
	now               int64 // precision: 10s

	counter *RequestCounter
	utils.Closable
}

func NewPrometheusLabelTable(controllerIPs []string, port, rpcMaxMsgSize, cacheExpiration int) *PrometheusLabelTable {
	ips := make([]net.IP, len(controllerIPs))
	for i, ipString := range controllerIPs {
		ips[i] = net.ParseIP(ipString)
		if ips[i].To4() != nil {
			ips[i] = ips[i].To4()
		}
	}
	t := &PrometheusLabelTable{
		GrpcSession:     &grpc.GrpcSession{},
		cacheExpiration: cacheExpiration,
		now:             time.Now().Unix(),
		counter:         &RequestCounter{},
	}
	for i := 0; i < MAX_ORG_COUNT; i++ {
		t.metricNameIDs[i] = hashmap.New[string, uint64]()     // metricName => metricID
		t.labelNameIDs[i] = hashmap.New[string, uint64]()      // labelName  => labelNameID
		t.labelValueIDs[i] = hashmap.New[string, uint64]()     // labelValue => labelValueID
		t.labelNameValues[i] = hashmap.New[uint64, struct{}]() // labelNameValue => exists
		t.labelColumnIndexs[i] = hashmap.New[uint64, uint32]() // metricID + LabelNameID => columnIndex
	}

	t.GrpcSession.Init(ips, uint16(port), grpc.DEFAULT_SYNC_INTERVAL, rpcMaxMsgSize, nil)
	log.Infof("New PrometheusLabelTable ips:%v port:%d rpcMaxMsgSize:%d", ips, port, rpcMaxMsgSize)
	debug.ServerRegisterSimple(ingesterctl.CMD_PROMETHEUS_LABEL, t)
	common.RegisterCountableForIngester("prometheus-label-request", t)
	return t
}

func (t *PrometheusLabelTable) DropOrg(orgId uint16) {
	t.metricNameIDs[orgId] = hashmap.New[string, uint64]()
	t.labelNameIDs[orgId] = hashmap.New[string, uint64]()
	t.labelValueIDs[orgId] = hashmap.New[string, uint64]()
	t.labelNameValues[orgId] = hashmap.New[uint64, struct{}]()
	t.labelColumnIndexs[orgId] = hashmap.New[uint64, uint32]()
}

func (t *PrometheusLabelTable) RequestLabelIDs(request *trident.PrometheusLabelRequest) (*trident.PrometheusLabelResponse, error) {
	t.counter.RequestCount++
	t.counter.RequestLabelsCount += int64(len(request.GetRequestLabels()))
	var response *trident.PrometheusLabelResponse
	requestStart := time.Now()
	err := t.GrpcSession.Request(func(ctx context.Context, remote net.IP) error {
		var err error
		c := t.GrpcSession.GetClient()
		if c == nil {
			return fmt.Errorf("can't get grpc client to %s", remote)
		}
		client := trident.NewSynchronizerClient(c)
		response, err = client.GetPrometheusLabelIDs(ctx, request)
		return err
	})
	if err != nil {
		t.counter.ResponseFailed++
		return nil, err
	}

	t.counter.ResponseLabelsCount += int64(len(response.GetResponseLabelIds()))
	t.counter.RequestTotalDelayNs += int64(time.Since(requestStart))
	isAll := false
	if len(request.RequestLabels) == 0 && len(request.RequestTargets) == 0 {
		isAll = true
	}
	t.updatePrometheusLabels(response, isAll)

	return response, nil
}

func (t *PrometheusLabelTable) RequestAllLabelIDs() {
	log.Info("prometheus request all label IDs start")
	_, err := t.RequestLabelIDs(&trident.PrometheusLabelRequest{})
	if err != nil {
		log.Warning("request all prometheus label ids failed: %s", err)
	}
	log.Infof("prometheus request all label IDs end. %s", t.statsString(ckdb.DEFAULT_ORG_ID))
}

func u32SliceIsEqual(l, r []uint32) bool {
	if len(l) != len(r) {
		return false
	}
	for i := range l {
		if l[i] != r[i] {
			return false
		}
	}
	return true
}

// get elements that appear only once in the two slices
func uniqueElements(slice1, slice2 []uint32) map[uint32]bool {
	unique := make(map[uint32]bool)

	for _, v1 := range slice1 {
		unique[v1] = true
	}

	for _, v2 := range slice2 {
		if _, ok := unique[v2]; !ok {
			unique[v2] = true
		} else {
			delete(unique, v2)
		}
	}

	return unique
}

func (t *PrometheusLabelTable) genId(isAll bool, id uint32) uint64 {
	if isAll {
		// set expiration time evenly by ID to preven expiration at same time
		return uint64(id)<<32 | (uint64(t.now) - (uint64(id) % uint64(t.cacheExpiration)))
	}
	return uint64(id)<<32 | uint64(t.now)
}

func (t *PrometheusLabelTable) getId(value uint64) (id uint32, valid bool) {
	timestamp := uint32(value)
	if t.now-int64(timestamp) > int64(t.cacheExpiration) {
		t.counter.CacheExpiration++
		return uint32(value >> 32), false
	}
	return uint32(value >> 32), true
}

func (t *PrometheusLabelTable) updatePrometheusLabels(resp *trident.PrometheusLabelResponse, isAll bool) {
	if isAll {
		for _, orgLabelInfos := range resp.GetOrgResponseLabels() {
			orgId := orgLabelInfos.GetOrgId()
			for _, labelInfo := range orgLabelInfos.GetResponseLabels() {
				name := labelInfo.GetName()
				nameId := labelInfo.GetNameId()
				if name != "" && nameId != 0 {
					t.labelNameIDs[orgId].Set(strings.Clone(name), t.genId(isAll, nameId))
				} else {
					t.counter.LabelNameUnknown++
				}
				value := labelInfo.GetValue()
				valueId := labelInfo.GetValueId()
				if valueId != 0 {
					t.labelValueIDs[orgId].Set(strings.Clone(value), t.genId(isAll, valueId))
				} else {
					t.counter.LabelValueUnknown++
				}
				t.labelNameValues[orgId].Set(nameValueKey(nameId, valueId), struct{}{})
			}
		}
	}

	for _, metric := range resp.GetResponseLabelIds() {
		metricName := metric.GetMetricName()
		if metricName == "" {
			t.counter.MetricUnknown++
			continue
		}
		orgId := metric.GetOrgId()
		metricId := metric.GetMetricId()
		t.metricNameIDs[orgId].Set(strings.Clone(metricName), t.genId(isAll, metricId))
		for _, labelInfo := range metric.GetLabelIds() {
			name := labelInfo.GetName()
			nameId := labelInfo.GetNameId()
			if name != "" && nameId != 0 {
				t.labelNameIDs[orgId].Set(strings.Clone(name), t.genId(isAll, nameId))
			} else {
				t.counter.LabelNameUnknown++
			}
			// if get all lables, value info is nothing
			if !isAll {
				value := labelInfo.GetValue()
				valueId := labelInfo.GetValueId()
				if valueId != 0 {
					t.labelValueIDs[orgId].Set(strings.Clone(value), t.genId(isAll, valueId))
				} else {
					t.counter.LabelValueUnknown++
				}
				t.labelNameValues[orgId].Set(nameValueKey(nameId, valueId), struct{}{})
			}

			cIndex := labelInfo.GetAppLabelColumnIndex()
			t.labelColumnIndexs[orgId].Set(columnIndexKey(metricId, nameId), cIndex)
		}
	}
}

func (t *PrometheusLabelTable) GetMaxAppLabelColumnIndex() int {
	ret := 0
	for _, labelColumnIndex := range t.labelColumnIndexs {
		maxIndex := int(getUInt64MapMaxValue(labelColumnIndex))
		if maxIndex > ret {
			ret = maxIndex
		}
	}
	return ret
}

func (t *PrometheusLabelTable) metricIDsString(orgId uint16, filter string) string {
	sb := &strings.Builder{}
	sb.WriteString(fmt.Sprintf("\norg-id %d\n", orgId))
	sb.WriteString("\nmetricName                                                                                            metricId   updated_at\n")
	sb.WriteString("------------------------------------------------------------------------------------------------------------------------------\n")
	t.metricNameIDs[orgId].Range(func(k string, v uint64) bool {
		row := fmt.Sprintf("%-100s  %-8d   %s\n", k, v>>32, time.Unix(int64(v<<32>>32), 0).Format("2006-01-02T15:04:05Z"))
		if strings.Contains(row, filter) {
			sb.WriteString(row)
		}
		return true
	})
	return sb.String()
}

func (t *PrometheusLabelTable) nameIDsString(orgId uint16, filter string) string {
	sb := &strings.Builder{}
	sb.WriteString(fmt.Sprintf("\norg-id %d\n", orgId))
	sb.WriteString("\nname                                                              nameId    updated_at\n")
	sb.WriteString("----------------------------------------------------------------------------------------\n")
	t.labelNameIDs[orgId].Range(func(k string, v uint64) bool {
		row := fmt.Sprintf("%-64s  %-7d   %s\n", k, v>>32, time.Unix(int64(v<<32>>32), 0).Format("2006-01-02T15:04:05Z"))
		if strings.Contains(row, filter) {
			sb.WriteString(row)
		}
		return true
	})
	return sb.String()
}

func (t *PrometheusLabelTable) valueIDsString(orgId uint16, filter string) string {
	sb := &strings.Builder{}
	sb.WriteString(fmt.Sprintf("\norg-id %d\n", orgId))
	sb.WriteString("\nvalue                                                                                                                             valueId    updated_at\n")
	sb.WriteString("--------------------------------------------------------------------------------------------------------------------------------------------------------\n")
	t.labelValueIDs[orgId].Range(func(k string, v uint64) bool {
		row := fmt.Sprintf("%-128s  %-8d   %s\n", k, v>>32, time.Unix(int64(v<<32>>32), 0).Format("2006-01-02T15:04:05Z"))
		if strings.Contains(row, filter) {
			sb.WriteString(row)
		}
		return true
	})
	return sb.String()
}

func (t *PrometheusLabelTable) columnIndexString(orgId uint16, filter string) string {
	sb := &strings.Builder{}
	sb.WriteString(fmt.Sprintf("\norg-id %d\n", orgId))
	sb.WriteString("\ncolumnIndex  metricName                                                                                            metricId   name                                                              nameId\n")
	sb.WriteString("--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------\n")
	t.labelColumnIndexs[orgId].Range(func(k uint64, v uint32) bool {
		metricId := k >> METRICID_OFFSET
		nameId := k << (64 - METRICID_OFFSET) >> (64 - METRICID_OFFSET)
		metricName, name := "", ""
		t.metricNameIDs[orgId].Range(func(n string, i uint64) bool {
			if i>>32 == metricId {
				metricName = n
				return false
			}
			return true
		})
		t.labelNameIDs[orgId].Range(func(n string, i uint64) bool {
			if i>>32 == nameId {
				name = n
				return false
			}
			return true
		})
		row := fmt.Sprintf("%-11d  %-100s  %-9d  %-64s  %-6d\n", v, metricName, metricId, name, nameId)
		if strings.Contains(row, filter) {
			sb.WriteString(row)
		}
		return true
	})
	return sb.String()
}

func getStringMapMaxValue(m *hashmap.Map[string, uint64]) uint32 {
	maxId := uint32(0)
	m.Range(func(n string, i uint64) bool {
		if uint32(i>>32) > maxId {
			maxId = uint32(i >> 32)
		}
		return true
	})
	return maxId
}

func getUInt64MapMaxValue(m *hashmap.Map[uint64, uint32]) uint32 {
	maxId := uint32(0)
	m.Range(func(n uint64, i uint32) bool {
		if i > maxId {
			maxId = i
		}
		return true
	})
	return maxId
}

func getArrayUInt64MapMaxValue(m *hashmap.Map[complex128, uint32]) uint32 {
	maxId := uint32(0)
	m.Range(func(_ complex128, i uint32) bool {
		if i > maxId {
			maxId = i
		}
		return true
	})
	return maxId
}

func (t *PrometheusLabelTable) statsString(orgId uint16) string {
	sb := &strings.Builder{}
	sb.WriteString(fmt.Sprintf("\norg-id %d\n", orgId))
	sb.WriteString("\ntableType  total-count  max-id\n")
	sb.WriteString("--------------------------------\n")
	sb.WriteString(fmt.Sprintf("%-9s  %-11d  %-6d\n", "metric", t.metricNameIDs[orgId].Len(), getStringMapMaxValue(t.metricNameIDs[orgId])))
	sb.WriteString(fmt.Sprintf("%-9s  %-11d  %-6d\n", "name", t.labelNameIDs[orgId].Len(), getStringMapMaxValue(t.labelNameIDs[orgId])))
	sb.WriteString(fmt.Sprintf("%-9s  %-11d  %-6d\n", "value", t.labelValueIDs[orgId].Len(), getStringMapMaxValue(t.labelValueIDs[orgId])))
	sb.WriteString(fmt.Sprintf("%-9s  %-11d  %-6d\n", "column", t.labelColumnIndexs[orgId].Len(), getUInt64MapMaxValue(t.labelColumnIndexs[orgId])))
	return sb.String()
}

func (t *PrometheusLabelTable) HandleSimpleCommand(op uint16, args string) string {
	parts := strings.Split(args, "|")
	id, _ := strconv.Atoi(parts[0])
	orgId := uint16(id)
	filter := ""
	if len(parts) > 1 {
		filter = parts[1]
	}
	cmd := labelCmds[op]
	switch cmd {
	case "metric":
		return t.metricIDsString(orgId, filter)
	case "name":
		return t.nameIDsString(orgId, filter)
	case "value":
		return t.valueIDsString(orgId, filter)
	case "column":
		return t.columnIndexString(orgId, filter)
	case "stats":
		return t.statsString(orgId)
	case "test":
		return t.testString(args)
	case "explain":
		return t.explainString(args)
	}
	return t.statsString(orgId)
}

// request string as: metric=xxx,pod_cluster_id=xxx,epc_id=xxx,label1=xxx,label2=xxx
func (t *PrometheusLabelTable) testString(request string) string {
	req := &trident.PrometheusLabelRequest{}
	metricReq := &trident.MetricLabelRequest{}
	targetReq := &trident.TargetRequest{}
	keyValues := strings.Split(request, ",")
	clusterId, epcId, orgId := 0, 0, 0
	for _, kv := range keyValues {
		kv := strings.Split(kv, "=")
		if len(kv) != 2 {
			continue
		}
		if kv[0] == "metric" {
			metricReq.MetricName = &(kv[1])
		} else if kv[0] == "pod_cluster_id" {
			clusterId, _ = strconv.Atoi(kv[1])
		} else if kv[0] == "epc_id" {
			epcId, _ = strconv.Atoi(kv[1])
		} else if kv[0] == "org_id" {
			orgId, _ = strconv.Atoi(kv[1])
		} else {
			addLabel(metricReq, kv[0], kv[1])
		}
	}
	metricReq.PodClusterId = proto.Uint32(uint32(clusterId))
	metricReq.EpcId = proto.Uint32(uint32(epcId))
	targetReq.PodClusterId = proto.Uint32(uint32(clusterId))
	targetReq.EpcId = proto.Uint32(uint32(epcId))
	targetReq.OrgId = proto.Uint32(uint32(orgId))
	req.RequestLabels = append(req.RequestLabels, metricReq)
	req.RequestTargets = append(req.RequestTargets, targetReq)
	if request == "all" {
		req = &trident.PrometheusLabelRequest{}
	}
	resp, err := t.RequestLabelIDs(req)
	if err != nil {
		return fmt.Sprintf("request: %s\nresponse failed: %s", req, err)
	}
	return fmt.Sprintf("request: %s\nresponse: %s", req, resp)
}

// explain string as: xxx|xxx|xxxx|xxxx|...,   means: metric_id|app_label_value_id_1|app_label_value_id_2|...
func (t *PrometheusLabelTable) explainString(str string) string {
	values := strings.Split(str, "|")
	intValues := []int{}
	for _, v := range values {
		i := strings.TrimSpace(v)
		if len(i) > 0 {
			integer, _ := strconv.Atoi(i)
			intValues = append(intValues, integer)
		}
	}
	orgId := 1
	explainStr, explainedStr := "", ""
	for i, v := range intValues {
		if i == 0 {
			explainStr += fmt.Sprintf("metric_id=%d,", v)
		} else {
			explainStr += fmt.Sprintf("app_label_value_id_%d=%d,", i-1, v)
		}
	}
	metricName := ""
	if len(intValues) < 1 || intValues[0] == 0 {
		return fmt.Sprintf("invalid metric_id, %s", explainStr)
	}
	metricId := intValues[0]
	t.metricNameIDs[orgId].Range(func(n string, i uint64) bool {
		if i>>32 == uint64(metricId) {
			metricName = n
			return false
		}
		return true
	})

	if len(intValues) < 2 || intValues[1] == 0 {
		return fmt.Sprintf("invalid target_id, %s", explainStr)
	}

	names, values := make([]string, len(intValues)-1), make([]string, len(intValues)-1)

	for i, valueId := range intValues[2:] {
		t.labelColumnIndexs[orgId].Range(func(k uint64, v uint32) bool {
			mid := k >> METRICID_OFFSET
			if uint64(metricId) != mid {
				return true
			}
			if v == uint32(i+1) {
				nameId := k << (64 - METRICID_OFFSET) >> (64 - METRICID_OFFSET)
				name, value := "", ""
				t.labelNameIDs[orgId].Range(func(n string, i uint64) bool {
					if i>>32 == nameId {
						name = n
						return false
					}
					return true
				})
				t.labelValueIDs[orgId].Range(func(n string, i uint64) bool {
					if i>>32 == uint64(valueId) {
						value = n
						return false
					}
					return true
				})
				names[v] = name
				values[v] = value
				return false
			}
			return true
		})

	}

	explainedStr += fmt.Sprintf("metric=%s,", metricName)
	for i := range names {
		if i == 0 {
			continue
		}
		explainedStr += fmt.Sprintf("[%d]%s=%s,", i, names[i], values[i])
	}

	return fmt.Sprintf("explain: %s\nexplained: %s", explainStr, explainedStr)
}

var labelCmds = []string{"metric", "name", "value", "column", "stats", "test", "explain"}
var cmdHelps = []string{"[filter]", "[filter]", "[filter]", "[filter]", "", "metric=xxx,label1=xxx,label2=xxx", "xxx|xxx|xxxx|xxxx|..., means: metric_id|target_id|app_label_value_id_1|app_label_value_id_2|..."}

func RegisterClientPrometheusLabelCommand() *cobra.Command {
	operates := []debug.CmdHelper{}
	for i, cmd := range labelCmds {
		operates = append(operates, debug.CmdHelper{Cmd: cmd, Helper: cmdHelps[i]})
	}

	return debug.ClientRegisterSimple(ingesterctl.CMD_PROMETHEUS_LABEL,
		debug.CmdHelper{
			Cmd:    "label",
			Helper: "show prometheus label info",
		},
		operates,
	)
}
