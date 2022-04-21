use super::super::{
    consts::*, AppProtoHead, AppProtoLogsInfo, L7LogParse, L7Protocol, L7ResponseStatus,
    LogMessageType,
};

use crate::common::enums::{IpProtocol, PacketDirection};
use crate::flow_generator::error::{Error, Result};
use crate::proto::flow_log;
use crate::utils::bytes::{read_u32_be, read_u64_be};

#[derive(Debug, Default, Clone)]
pub struct DubboInfo {
    // header
    pub serial_id: u8,
    pub data_type: u8,
    pub request_id: i64,

    // req
    pub req_msg_size: i32,
    pub dubbo_version: String,
    pub service_name: String,
    pub service_version: String,
    pub method_name: String,

    // resp
    pub resp_msg_size: i32,
}

impl DubboInfo {
    pub fn merge(&mut self, other: Self) {
        self.resp_msg_size = other.resp_msg_size;
    }
}

impl From<DubboInfo> for flow_log::DubboInfo {
    fn from(f: DubboInfo) -> Self {
        flow_log::DubboInfo {
            serial_id: f.serial_id as u32,
            r#type: f.data_type as u32,
            id: f.request_id as u32,
            req_body_len: f.req_msg_size,
            version: f.dubbo_version,
            service_name: f.service_name,
            service_version: f.service_version,
            method_name: f.method_name,
            resp_body_len: f.resp_msg_size,
            trace_id: Default::default(),
        }
    }
}

#[derive(Clone, Debug, Default)]
pub struct DubboLog {
    info: DubboInfo,

    status: L7ResponseStatus,
    status_code: u8,
    msg_type: LogMessageType,
}

impl DubboLog {
    fn reset_logs(&mut self) {
        self.info.serial_id = 0;
        self.info.data_type = 0;
        self.info.request_id = 0;
        self.info.req_msg_size = -1;
        self.info.dubbo_version = String::new();
        self.info.service_name = String::new();
        self.info.service_version = String::new();
        self.info.method_name = String::new();
        self.info.resp_msg_size = -1;
    }

    // 尽力而为的去解析Dubbo请求中Body各参数
    fn get_req_body_info(&mut self, payload: &[u8]) {
        let mut n = BODY_PARAM_MIN;
        let mut para_index = 1;
        let mut para_tag = payload[0];
        let payload_len = payload.len();

        while n < BODY_PARAM_MAX {
            let para_len = match get_req_param_len(para_tag) {
                Some(len) if payload_len >= para_index + len => len,
                _ => return,
            };

            match n {
                BODY_PARAM_DUBBO_VERSION => {
                    self.info.dubbo_version =
                        String::from_utf8_lossy(&payload[para_index..para_index + para_len])
                            .into_owned()
                }

                BODY_PARAM_SERVICE_NAME => {
                    self.info.service_name =
                        String::from_utf8_lossy(&payload[para_index..para_index + para_len])
                            .into_owned();
                }

                BODY_PARAM_SERVICE_VERSION => {
                    self.info.service_version =
                        String::from_utf8_lossy(&payload[para_index..para_index + para_len])
                            .into_owned();
                }

                BODY_PARAM_METHOD_NAME => {
                    self.info.method_name =
                        String::from_utf8_lossy(&payload[para_index..para_index + para_len])
                            .into_owned();
                }
                _ => return,
            }

            para_index += para_len;
            if payload_len <= para_index {
                return;
            }
            para_tag = payload[para_index];
            para_index += 1;
            n += 1;
        }
    }

    fn request(&mut self, payload: &[u8], dubbo_header: &DubboHeader) {
        self.msg_type = LogMessageType::Request;

        self.info.data_type = dubbo_header.data_type;
        self.info.req_msg_size = dubbo_header.data_length;
        self.info.serial_id = dubbo_header.serial_id;
        self.info.request_id = dubbo_header.request_id;

        self.get_req_body_info(&payload[DUBBO_HEADER_LEN..]);
    }

    fn set_status(&mut self, status_code: u8) {
        self.status = match status_code {
            20 => L7ResponseStatus::Ok,
            30 => L7ResponseStatus::ClientError,
            31 => L7ResponseStatus::ServerError,
            40 => L7ResponseStatus::ClientError,
            50 => L7ResponseStatus::ServerError,
            60 => L7ResponseStatus::ServerError,
            70 => L7ResponseStatus::ServerError,
            80 => L7ResponseStatus::ServerError,
            90 => L7ResponseStatus::ClientError,
            100 => L7ResponseStatus::ServerError,
            _ => L7ResponseStatus::Ok,
        }
    }

    fn response(&mut self, dubbo_header: &DubboHeader) {
        self.msg_type = LogMessageType::Response;

        self.info.data_type = dubbo_header.data_type;
        self.info.resp_msg_size = dubbo_header.data_length;
        self.info.serial_id = dubbo_header.serial_id;
        self.info.request_id = dubbo_header.request_id;
        self.status_code = dubbo_header.status_code;
        self.set_status(self.status_code);
    }
}

impl L7LogParse for DubboLog {
    fn parse(
        &mut self,
        payload: &[u8],
        proto: IpProtocol,
        direction: PacketDirection,
    ) -> Result<AppProtoHead> {
        if proto != IpProtocol::Tcp {
            return Err(Error::InvalidIpProtocol);
        }

        self.reset_logs();
        let mut dubbo_header = DubboHeader::default();
        dubbo_header.parse_headers(payload)?;

        match direction {
            PacketDirection::ClientToServer => {
                self.request(payload, &dubbo_header);
            }
            PacketDirection::ServerToClient => {
                self.response(&dubbo_header);
            }
        }
        Ok(AppProtoHead {
            proto: L7Protocol::Dubbo,
            msg_type: self.msg_type,
            status: self.status,
            code: self.status_code as u16,
            rrt: 0,
        })
    }

    fn info(&self) -> AppProtoLogsInfo {
        AppProtoLogsInfo::Dubbo(self.info.clone())
    }
}

#[derive(Debug, Default, PartialEq)]
pub struct DubboHeader {
    // Dubbo Header
    pub serial_id: u8,
    pub data_type: u8,
    pub status_code: u8,
    pub data_length: i32,
    pub request_id: i64,
}

impl DubboHeader {
    // Dubbo协议https://dubbo.apache.org/zh/blog/2018/10/05/dubbo-%E5%8D%8F%E8%AE%AE%E8%AF%A6%E8%A7%A3/#dubbo-%E5%8D%8F%E8%AE%AE
    // Dubbo协议帧
    // +-----------------------------------------------+
    // |       header           |       body           |
    // +---------------+---------------+---------------+
    // header格式
    // +------------------------------------------------------------------------------------------------------------+
    // | magic (16) | request and serialization flag (8) | response status (8) | request id (64) | body length (32) |
    // +------------------------------------------------------------------------------------------------------------+
    pub fn parse_headers(&mut self, payload: &[u8]) -> Result<()> {
        if payload.len() < DUBBO_HEADER_LEN {
            return Err(Error::DubboHeaderParseFailed);
        }
        if payload[0] != DUBBO_MAGIC_HIGH || payload[1] != DUBBO_MAGIC_LOW {
            return Err(Error::DubboHeaderParseFailed);
        }

        self.serial_id = payload[2] & 0x1f;
        self.data_type = payload[2] & 0x80;
        self.status_code = payload[3];
        self.request_id = read_u64_be(&payload[4..]) as i64;
        self.data_length = read_u32_be(&payload[12..]) as i32;
        Ok(())
    }
}

// 参考开源代码解析：https://github.com/apache/dubbo-go-hessian2/blob/master/decode.go#L289
pub fn get_req_param_len(tag: u8) -> Option<usize> {
    if (tag == BC_STRING_CHUNK || tag == BC_STRING)
        || (tag >= BC_STRING_DIRECT && tag <= STRING_DIRECT_MAX)
        || (tag >= 0x30 && tag <= 0x33)
    {
        return Some(tag as usize);
    }
    None
}

#[cfg(test)]
mod tests {
    use std::fs;
    use std::path::Path;

    use super::*;

    use crate::{common::enums::PacketDirection, utils::test::Capture};

    const FILE_DIR: &str = "resources/test/flow_generator/dubbo";

    fn run(name: &str) -> String {
        let capture = Capture::load_pcap(Path::new(FILE_DIR).join(name), None);
        let mut packets = capture.as_meta_packets();
        if packets.is_empty() {
            return "".to_string();
        }

        let mut output: String = String::new();
        let first_dst_port = packets[0].lookup_key.dst_port;
        for packet in packets.iter_mut() {
            packet.direction = if packet.lookup_key.dst_port == first_dst_port {
                PacketDirection::ClientToServer
            } else {
                PacketDirection::ServerToClient
            };
            let payload = match packet.get_l4_payload() {
                Some(p) => p,
                None => continue,
            };

            let mut dubbo = DubboLog::default();
            let _ = dubbo.parse(payload, packet.lookup_key.proto, packet.direction);
            output.push_str(&format!("{:?}\r\n", dubbo.info));
        }
        output
    }

    #[test]
    fn check() {
        let files = vec![("dubbo_hessian2.pcap", "dubbo_hessian.result")];

        for item in files.iter() {
            let expected = fs::read_to_string(&Path::new(FILE_DIR).join(item.1)).unwrap();
            let output = run(item.0);

            if output != expected {
                let output_path = Path::new("actual.txt");
                fs::write(&output_path, &output).unwrap();
                assert!(
                    output == expected,
                    "output different from expected {}, written to {:?}",
                    item.1,
                    output_path
                );
            }
        }
    }
}
