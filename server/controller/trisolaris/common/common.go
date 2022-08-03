/*
 * Copyright (c) 2022 Yunshan Networks
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

package common

const (
	// VTAP
	VTAP_CONTROLLER_EXCEPTIONS_MASK = 0xFFFFFFFF00000000
	VTAP_TRIDENT_EXCEPTIONS_MASK    = 0x00000000FFFFFFFF
	VTAP_NO_REGISTER_EXCEPTION      = 0x20000000

	VTAP_TYPE_HYPER_V_NETWORK = 11

	SHUT_DOWN_STR  = "关闭"
	SHUT_DOWN_UINT = 0xffffffff

	ALL_CLUSTERS    = 0
	CLUSTER_OF_VTAP = 1

	DROPLET_PLATFORM_DATA               = 0
	ALL_SIMPLE_PLATFORM_DATA            = 1
	ALL_SIMPLE_PLATFORM_DATA_EXCEPT_POD = 2
	DOMAIN_TO_ALL_SIMPLE_PLATFORM_DATA  = 3
	DOMAIN_TO_PLATFORM_DATA_EXCEPT_POD  = 4
	DOMAIN_TO_PLATFORM_DATA_ONLY_POD    = 5
	ALL_SKIP_SIMPLE_PLATFORM_DATA       = 6
	PLATFORM_DATA_TYPE_1                = 6
	PLATFORM_DATA_TYPE_2                = 7
	PLATFORM_DATA_TYPE_3                = 8
	PLATFORM_DATA_BM_DEDICATED          = 9

	SKIP_PLATFORM_DATA_TYPE_1               = 10
	SKIP_PLATFORM_DATA_TYPE_2               = 11
	SKIP_PLATFORM_DATA_TYPE_3               = 12
	DOMAIN_TO_SKIP_ALL_SIMPLE_PLATFORM_DATA = 13
	DOMAIN_TO_SKIP_PLATFORM_DATA_EXCEPT_POD = 14
	DOMAIN_TO_SKIP_PLATFORM_DATA_ONLY_POD   = 15

	DEFAULT_MAX_MEMORY         = 256
	DEFAULT_MAX_ESCAPE_SECONDS = 3600

	CONTROLLER_VTAP_MAX = 2000
	TSDB_VTAP_MAX       = 200

	// TRIDENT OS
	TRIDENT_LINUX   = 0
	TRIDENT_WINDOWS = 1

	TSDB_PROCESS_NAME = "roze"

	CONN_DEFAULT_AZ     = "ALL"
	CONN_DEFAULT_REGION = "ffffffff-ffff-ffff-ffff-ffffffffffff"

	TAPMODE_LOCAL    = 0
	TAPMODE_MIRROR   = 1
	TAPMODE_ANALYZER = 2
	TAPMODE_DECAP    = 3

	NPB_BUSINESS_ID  = 1
	PCAP_BUSINESS_ID = -3

	RESOURCE_GROUP_TYPE_ANONYMOUS_POD_GROUP                = 6
	RESOURCE_GROUP_TYPE_VM                                 = 1
	RESOURCE_GROUP_TYPE_ANONYMOUS_VM                       = 3
	RESOURCE_GROUP_TYPE_ANONYMOUS_POD_SERVICE              = 8
	RESOURCE_GROUP_TYPE_ANONYMOUS_POD_GROUP_AS_POD_SERVICE = 81
	RESOURCE_GROUP_TYPE_ANONYMOUS_VL2                      = 14

	POD_SERVICE_TYPE_NODE_PORT = 2

	INTERNET_RESOURCE_GROUP_ID_UINT32 = -2 & 0xffffffff
	INTERNET_EPC_ID_UINT32            = -2 & 0xffffffff
	RESOURCE_GROUP_TYPE_NONE          = 0
	RESOURCE_GROUP_TYPE_ANONYMOUS_IP  = 4

	APPLICATION_ANALYSIS          = 1
	APPLICATION_FLOW_BACKTRACKING = 2
	APPLICATION_NPB               = 6
	APPLICATION_PCAP              = 4

	ALL_VTAP_SHARE_POLICY_VERSION_OFFSET = 1000000000
	DROPLET_POLICY_VERSION_OFFSET        = 2000000000

	PROTOCOL_ALL = 256

	MAX_PAYLOAD_SLICE = 65535

	BILLING_METHOD_LICENSE = "license"
	BILLING_METHOD_VOUCHER = "voucher"
)
