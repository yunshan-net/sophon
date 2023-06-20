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

mod debug;
mod overwrite_queue;

pub use debug::{bounded_with_debug, DebugSender};
pub use overwrite_queue::{bounded, Counter, Receiver, Sender, StatsHandle};

#[derive(Debug, PartialEq)]
pub enum Error<T> {
    Timeout,
    Terminated(Option<T>, Option<Vec<T>>),
    BatchTooLarge(Option<Vec<T>>),
}
