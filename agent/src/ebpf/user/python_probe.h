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

#ifndef _BPF_PYTHON_PROBE_H_
#define _BPF_PYTHON_PROBE_H_

#include "tracer.h"

// Scan /proc/ to get all processes when the agent starts;
int collect_python_uprobe_syms_from_procfs(struct tracer_probes_conf *conf);

void python_process_exec(int pid);

void python_process_exit(int pid);

void python_events_handle(void);

#endif
