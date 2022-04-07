use std::convert::TryInto;
use std::net::IpAddr;
use chrono::prelude::DateTime;
use chrono::Utc;
use chrono::FixedOffset;
use std::time::{UNIX_EPOCH, Duration};
use rust_sample::ebpf::*;
use std::fmt::Write;
use std::thread;

extern "C" {
    fn print_dns_info(data: *mut c_char, len: c_uint);
}

fn flow_info(sd: *mut SK_BPF_DATA) -> String {
    unsafe {
        let mut flow = String::from("");
        if (*sd).direction == SOCK_DIR_SND {
            write!(
                flow,
                "{} {}.{} > {}.{}",
                sk_l4proto_safe(sd),
                sk_laddr_str_safe(sd),
                (*sd).tuple.lport,
                sk_raddr_str_safe(sd),
                (*sd).tuple.rport
            )
            .unwrap();
        } else {
            write!(
                flow,
                "{} {}.{} > {}.{}",
                sk_l4proto_safe(sd),
                sk_raddr_str_safe(sd),
                (*sd).tuple.rport,
                sk_laddr_str_safe(sd),
                (*sd).tuple.lport
            )
            .unwrap();
        }

        return flow;
    }
}

fn date_time(ts: u64) -> String {
    // Creates a new SystemTime from the specified number of whole seconds
    let d = UNIX_EPOCH + Duration::from_micros(ts);
    // Create DateTime from SystemTime
    let time = DateTime::<Utc>::from(d);
    let china_timezone = FixedOffset::east(8 * 3600);
    // Formats the combined date and time with the specified format string.
    time.with_timezone(&china_timezone).format("%Y-%m-%d %H:%M:%S.%6f").to_string()
}

fn sk_str_safe(data: *mut c_char) -> String {
    unsafe { CStr::from_ptr(data).to_string_lossy().into_owned() }
}

fn sk_bytes(data: *mut c_char, len: u32) -> &'static [u8] {
    unsafe {
        let slice = std::slice::from_raw_parts(data, len as usize);
        &*(slice as *const [c_char] as *const [u8])
    }
}

fn sk_bytes_safe(data: *mut c_char, len: u32) -> Vec<u8> {
    sk_bytes(data, len).iter().cloned().collect()
}

fn sk_data_str_safe(sd: *mut SK_BPF_DATA) -> String {
    unsafe { sk_str_safe((*sd).cap_data) }
}

fn sk_data_bytes_safe(sd: *mut SK_BPF_DATA) -> Vec<u8> {
    unsafe { sk_bytes_safe((*sd).cap_data, (*sd).cap_len) }
}

fn sk_proto_safe(sd: *mut SK_BPF_DATA) -> u16 {
    unsafe { (*sd).l7_protocal_hint }
}

//>= Rust 1.34
fn pop_4(barry: &[u8]) -> [u8; 4] {
    barry.try_into().expect("slice with incorrect length")
}

fn sk_ip_string_safe(addr: [u8; 16usize], addr_len: u8) -> String {
    let ret: String = String::from("");
    if addr_len == 4 {
        return IpAddr::from(pop_4(&addr[0..4])).to_string();
    } else if addr_len == 16 {
        return IpAddr::from(addr).to_string();
    }

    return ret;
}

fn sk_raddr_str_safe(sd: *mut SK_BPF_DATA) -> String {
    unsafe { sk_ip_string_safe((*sd).tuple.raddr, (*sd).tuple.addr_len) }
}

fn sk_laddr_str_safe(sd: *mut SK_BPF_DATA) -> String {
    unsafe { sk_ip_string_safe((*sd).tuple.laddr, (*sd).tuple.addr_len) }
}

fn sk_l4proto_safe(sd: *mut SK_BPF_DATA) -> &'static str {
    unsafe {
        if (*sd).tuple.protocol == 6 {
            return "TCP";
        } else if (*sd).tuple.protocol == 17 {
            return "UDP";
        }

        return "";
    }
}

fn process_name_safe(sd: *mut SK_BPF_DATA) -> String {
	unsafe {
		let v = &(*sd).process_name;
		String::from_utf8_lossy(v).to_string()
	}
}

extern "C" fn socket_trace_callback(sd: *mut SK_BPF_DATA) {
    unsafe {
        let mut proto_tag = String::from("");
        if sk_proto_safe(sd) == SOCK_DATA_HTTP1 {
            proto_tag.push_str("HTTP1");
        } else if sk_proto_safe(sd) == SOCK_DATA_HTTP2 {
            proto_tag.push_str("HTTP2");
        } else if sk_proto_safe(sd) == SOCK_DATA_DNS {
            proto_tag.push_str("DNS");
        } else if sk_proto_safe(sd) == SOCK_DATA_MYSQL {
            proto_tag.push_str("MYSQL");
        } else if sk_proto_safe(sd) == SOCK_DATA_REDIS {
            proto_tag.push_str("REDIS");
        } else if sk_proto_safe(sd) == SOCK_DATA_KAFKA {
            proto_tag.push_str("KAFKA");
        } else if sk_proto_safe(sd) == SOCK_DATA_DUBBO {
            proto_tag.push_str("DUBBO");
        }

        println!("+ --------------------------------- +");
        if sk_proto_safe(sd) == SOCK_DATA_HTTP1 {
            let data = sk_data_str_safe(sd);
            println!("{} <{}> RECONFIRM {} DIR {} PID {} THREAD_ID {} COMM {} {} LEN {} SYSCALL_LEN {} SOCKET_ID 0x{:x} PROXY_TRACE 0x{:x} THREAD_TRACE 0x{:x} TCP_SEQ {} DATA_SEQ {} TimeStamp {}\n{}", 
                     date_time((*sd).timestamp),
                     proto_tag,
                     (*sd).need_reconfirm,
                     (*sd).direction,
                     (*sd).process_id,
                     (*sd).thread_id,
		     process_name_safe(sd),
                     flow_info(sd),
                     (*sd).cap_len,
                     (*sd).syscall_len,
                     (*sd).socket_id,
                     (*sd).syscall_trace_id_call,
                     (*sd).syscall_trace_id_session,
                     (*sd).tcp_seq,
                     (*sd).cap_seq,
                     (*sd).timestamp,
                     data);
        } else {
            let data: Vec<u8> = sk_data_bytes_safe(sd);
            println!("{} <{}> RECONFIRM {} DIR {} PID {} THREAD_ID {} COMM {} {} LEN {} SYSCALL_LEN {} SOCKET_ID 0x{:x} PROXY_TRACE 0x{:x} THREAD_TRACE 0x{:x} TCP_SEQ {} DATA_SEQ {} TimeStamp {}",
                     date_time((*sd).timestamp),
                     proto_tag,
                     (*sd).need_reconfirm,
                     (*sd).direction,
                     (*sd).process_id,
                     (*sd).thread_id,
		     process_name_safe(sd),
                     flow_info(sd),
                     (*sd).cap_len,
                     (*sd).syscall_len,
                     (*sd).socket_id,
                     (*sd).syscall_trace_id_call,
                     (*sd).syscall_trace_id_session,
                     (*sd).tcp_seq,
                     (*sd).cap_seq,
                     (*sd).timestamp);
            if sk_proto_safe(sd) == SOCK_DATA_DNS {
                print_dns_info((*sd).cap_data, (*sd).cap_len);
            } else {
                for x in data.into_iter() {
                    let b = x as char;
                    print!("{0}", b);
                }
            }
            print!("\x1b[0m\n");
        }

        println!("+ --------------------------------- +\n");
    }
}

fn main() {
    let log_file = CString::new("/var/log/metaflow-ebpf.log".as_bytes()).unwrap();
    let log_file_c = log_file.as_c_str();
    unsafe {
	// 第一个参数空指针传递可以填写std::ptr::null()
        if bpf_tracer_init(log_file_c.as_ptr(), true) != 0 {
           println!("bpf_tracer_init() file:{:?} error", log_file);
           ::std::process::exit(1);
        }

        if running_socket_tracer(
            socket_trace_callback, /* 回调接口 rust -> C */
            1,                     /* 工作线程数，是指用户态有多少线程参与数据处理 */
            128,                    /* 内核共享内存占用的页框数量, 值为2的次幂。用于perf数据传递 */
            65536,                 /* 环形缓存队列大小，值为2的次幂。e.g: 2,4,8,16,32,64,128 */
            524288, /* 设置用于socket追踪的hash表项最大值，取决于实际场景中并发请求数量 */
            524288, /* 设置用于线程追踪会话的hash表项最大值，SK_BPF_DATA结构的syscall_trace_id_session关联这个哈希表 */
	    520000 /* socket map表项进行清理的最大阈值，当前map的表项数量超过这个值进行map清理操作 */
        ) != 0 {
            println!("running_socket_tracer() error.");
            ::std::process::exit(1);
        }

	bpf_tracer_finish();

        let stats = socket_tracer_stats();
        print!("{:#?}\n", stats);
    }

    thread::sleep(Duration::from_secs(5));
    unsafe {
        let stats = socket_tracer_stats();
        print!("{:#?}\n", stats);
	print!("start start ...\n");
        if tracer_start() != 0 {
            println!("tracer_start error");
        }
	print!("start start finish\n");
        let stats = socket_tracer_stats();
        print!("{:#?}\n", stats);
    }

    loop {
        thread::sleep(Duration::from_secs(5));
    }
}
