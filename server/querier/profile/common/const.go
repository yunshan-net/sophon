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

package common

const (
	SUCCESS            = "SUCCESS"
	FAIL               = "FAIL"
	INVALID_PARAMETERS = "INVALID_PARAMETERS"
	INVALID_POST_DATA  = "INVALID_POST_DATA"
	SERVER_ERROR       = "SERVER_ERROR"
)

const (
	DATABASE_PROFILE     = "profile"
	TABLE_PROFILE        = "in_process"
	PROFILE_LOCATION_STR = "profile_location_str"
	PROFILE_VALUE        = "profile_value"
)

const (
	HEADER_KEY_X_ORG_ID            = "X-Org-Id"
	LANGUAGE_TYPE_EBPF             = "eBPF"
	MAX_KERNEL_STACK_DEPTH_DEFAULT = -1
)

const (
	TAG_AGENT_ID   = "agent_id"
	TAG_PROCESS_ID = "process_id"
)
