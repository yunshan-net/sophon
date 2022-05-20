use std::{net::IpAddr, time::Duration};

use super::{perf::FlowPerf, FlowState, FLOW_METRICS_PEER_DST, FLOW_METRICS_PEER_SRC};
use crate::{
    common::{
        decapsulate::TunnelType,
        endpoint::EndpointData,
        enums::{EthernetType, PacketDirection, TapType},
        flow::FlowMetricsPeer,
        lookup_key::LookupKey,
        meta_packet::MetaPacket,
        policy::PolicyData,
        tagged_flow::TaggedFlow,
        TapPort,
    },
    proto::common::TridentType,
    utils::{hasher::jenkins64, net::MacAddr},
};

#[repr(u8)]
enum MatchMac {
    None,
    Dst,
    Src,
    All,
}

/*
    FlowTimeKey是用在时间集合流节点映射表的唯一标识。
    timestamp_key是一个时间戳的纳秒数，是每条flow的唯一时间标识，用作时间集合的排序，会随着流创建和定时刷新而变化。
    因为定时刷新和删除流都需要节点映射表的对应流节点,所以需要存一个FlowMapKey来找到节点对应的流slot，因为FlowMapKey用静态FlowKey，tap_port和tap_port生成,
    在复杂的网络环境可能相同，需要通过FlowNode.match_node方法找到对应的节点。
*/
#[derive(Default, Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord)]
pub(super) struct FlowTimeKey {
    // 作为时间集合的标识
    // 超过这个时间2554-07-22 07:34:33 GMT+0800 (中国标准时间)会溢出
    pub timestamp_key: u64,
    pub map_key: FlowMapKey,
}

impl FlowTimeKey {
    pub fn new(timestamp: Duration, map_key: FlowMapKey) -> Self {
        Self {
            timestamp_key: timestamp.as_nanos() as u64,
            map_key,
        }
    }

    pub fn reset_timestamp_key(&mut self, timestamp: u64) {
        self.timestamp_key = timestamp;
    }
}

/*
    FlowMapKey是流节点映射表的唯一标识,由Jenkins64算法哈希得到，因为FlowMap处理复杂网络环境，
    所以有可能key对应多个流节点的情况，需要根据流节点的match_node方法在映射表唯一标识一条流。
*/
#[derive(Debug, PartialEq, Eq, PartialOrd, Ord, Hash, Clone, Copy, Default)]
pub(super) struct FlowMapKey(pub u64);

impl FlowMapKey {
    fn l3_hash(lookup_key: &LookupKey) -> u64 {
        let (src, dst) = match (lookup_key.src_ip, lookup_key.dst_ip) {
            (IpAddr::V4(s), IpAddr::V4(d)) => (
                u32::from_le_bytes(s.octets()),
                u32::from_le_bytes(d.octets()),
            ),
            (IpAddr::V6(s), IpAddr::V6(d)) => {
                let (src, dst) = (s.octets(), d.octets());
                src.chunks(4)
                    .zip(dst.chunks(4))
                    .fold((0, 0), |(hash1, hash2), (b1, b2)| {
                        (
                            hash1 ^ u32::from_le_bytes(*<&[u8; 4]>::try_from(b1).unwrap()),
                            hash2 ^ u32::from_le_bytes(*<&[u8; 4]>::try_from(b2).unwrap()),
                        )
                    })
            }
            _ => unreachable!(),
        };

        if src >= dst {
            (src as u64) << 32 | dst as u64
        } else {
            (dst as u64) << 32 | src as u64
        }
    }

    fn l4_hash(lookup_key: &LookupKey) -> u64 {
        if lookup_key.src_port >= lookup_key.dst_port {
            (lookup_key.src_port as u64) << 16 | lookup_key.dst_port as u64
        } else {
            (lookup_key.dst_port as u64) << 16 | lookup_key.src_port as u64
        }
    }

    pub(super) fn new(lookup_key: &LookupKey, tap_port: TapPort) -> Self {
        match lookup_key.eth_type {
            EthernetType::Ipv4 | EthernetType::Ipv6 => {
                let lhs = Self::l3_hash(lookup_key);
                let rhs = ((u16::from(lookup_key.tap_type) as u64) << 24 | tap_port.0) << 32
                    | Self::l4_hash(lookup_key);
                Self(jenkins64(lhs) ^ jenkins64(rhs))
            }
            EthernetType::Arp => {
                let lhs = Self::l3_hash(lookup_key);
                let rhs = ((u16::from(lookup_key.tap_type) as u64) << 24 | tap_port.0 as u64) << 32
                    | (u64::from(lookup_key.src_mac) ^ u64::from(lookup_key.dst_mac));
                Self(jenkins64(lhs) ^ jenkins64(rhs))
            }
            _ => {
                let lhs = (u16::from(lookup_key.tap_type) as u64) << 24 | tap_port.0;
                let rhs = u64::from(lookup_key.src_mac) ^ u64::from(lookup_key.dst_mac);
                Self(jenkins64(lhs) ^ jenkins64(rhs))
            }
        }
    }
}

pub struct FlowNode {
    // 用作time_set比对的标识，等于FlowTimeKey的timestamp_key, 只有创建FlowNode和刷新更新流节点的超时才会更新
    pub timestamp_key: u64,

    pub tagged_flow: TaggedFlow,
    pub min_arrived_time: Duration,
    pub recent_time: Duration, // 最近一个Packet的时间戳
    pub timeout: Duration,     // 相对超时时间
    pub flow_state: FlowState,
    pub meta_flow_perf: Option<FlowPerf>,

    pub policy_data_cache: [PolicyData; 2],
    pub endpoint_data_cache: EndpointData,

    pub next_tcp_seq0: u32,
    pub next_tcp_seq1: u32,
    pub policy_in_tick: [bool; 2], // 当前统计周期（目前是自然秒）是否更新策略
    pub packet_in_tick: bool,      // 当前统计周期（目前是自然秒）是否有包
}

impl FlowNode {
    pub(super) fn reset_flow_stat_info(&mut self) {
        self.policy_in_tick = [false; 2];
        self.packet_in_tick = false;
        let flow = &mut self.tagged_flow.flow;
        flow.flow_stat_time = Duration::ZERO;
        flow.is_new_flow = false;
        let flow_metrics_peer_src = &mut flow.flow_metrics_peers[FLOW_METRICS_PEER_SRC];
        flow_metrics_peer_src.packet_count = 0;
        flow_metrics_peer_src.byte_count = 0;
        flow_metrics_peer_src.l3_byte_count = 0;
        flow_metrics_peer_src.l4_byte_count = 0;

        let flow_metrics_peer_dst = &mut flow.flow_metrics_peers[FLOW_METRICS_PEER_DST];
        flow_metrics_peer_dst.packet_count = 0;
        flow_metrics_peer_dst.byte_count = 0;
        flow_metrics_peer_dst.l3_byte_count = 0;
        flow_metrics_peer_dst.l4_byte_count = 0;
    }

    pub fn match_node(
        &self,
        meta_packet: &mut MetaPacket,
        config_ignore: (bool, bool),
        trident_type: TridentType,
    ) -> bool {
        let flow = &self.tagged_flow.flow;
        let flow_key = &flow.flow_key;
        let meta_lookup_key = &meta_packet.lookup_key;
        if flow_key.tap_port != meta_packet.tap_port
            || flow_key.tap_type != meta_lookup_key.tap_type
        {
            return false;
        }
        // other ethernet type
        if flow.eth_type != EthernetType::Ipv4 && meta_lookup_key.eth_type != EthernetType::Ipv6 {
            if meta_lookup_key.eth_type != flow.eth_type {
                return false;
            }
            // direction = ClientToServer
            if flow_key.mac_src == meta_lookup_key.src_mac
                && flow_key.mac_dst == meta_lookup_key.dst_mac
                && flow_key.ip_src == meta_lookup_key.src_ip
                && flow_key.ip_dst == meta_lookup_key.dst_ip
            {
                meta_packet.direction = PacketDirection::ClientToServer;
                return true;
            }
            // direction = ServerToClient
            if flow_key.mac_src == meta_lookup_key.dst_mac
                && flow_key.mac_dst == meta_lookup_key.src_mac
                && flow_key.ip_src == meta_lookup_key.dst_ip
                && flow_key.ip_dst == meta_lookup_key.src_ip
            {
                meta_packet.direction = PacketDirection::ServerToClient;
                return true;
            }

            return false;
        }

        if flow.eth_type != meta_lookup_key.eth_type {
            return false;
        }

        if flow_key.proto != meta_lookup_key.proto {
            return false;
        }

        if (meta_packet.tunnel.is_some()
            && flow.tunnel.tunnel_type != meta_packet.tunnel.unwrap().tunnel_type)
            || (meta_packet.tunnel.is_none() && flow.tunnel.tunnel_type != TunnelType::None)
        {
            // 微软ACS存在非对称隧道流量，需要排除
            if !Self::is_hyper_v(trident_type) {
                return false;
            }
        }

        // Ipv4/Ipv6 solve
        let mac_match = Self::mac_match(meta_packet, config_ignore, trident_type);
        if flow_key.ip_src == meta_lookup_key.src_ip
            && flow_key.ip_dst == meta_lookup_key.dst_ip
            && flow_key.port_src == meta_lookup_key.src_port
            && flow_key.port_dst == meta_lookup_key.dst_port
        {
            meta_packet.direction = PacketDirection::ClientToServer;
            Self::endpoint_match_with_direction(
                &flow.flow_metrics_peers,
                meta_packet,
                PacketDirection::ClientToServer,
            ) && Self::mac_match_with_direction(
                meta_packet,
                flow_key.mac_src,
                flow_key.mac_dst,
                mac_match,
                PacketDirection::ClientToServer,
            )
        } else if flow_key.ip_src == meta_lookup_key.dst_ip
            && flow_key.ip_dst == meta_lookup_key.src_ip
            && flow_key.port_src == meta_lookup_key.dst_port
            && flow_key.port_dst == meta_lookup_key.src_port
        {
            meta_packet.direction = PacketDirection::ServerToClient;
            Self::endpoint_match_with_direction(
                &flow.flow_metrics_peers,
                meta_packet,
                PacketDirection::ServerToClient,
            ) && Self::mac_match_with_direction(
                meta_packet,
                flow_key.mac_src,
                flow_key.mac_dst,
                mac_match,
                PacketDirection::ServerToClient,
            )
        } else {
            false
        }
    }

    fn is_hyper_v(trident_type: TridentType) -> bool {
        trident_type == TridentType::TtHyperVCompute || trident_type == TridentType::TtHyperVNetwork
    }

    // 微软ACS：
    //   HyperVNetwork网关宿主机和HyperVCompute网关流量模型中，MAC地址不对称
    //   在浦发环境中，IP地址不存在相同的场景，所以流聚合可直接忽略MAC地址
    //   但注意：若K8s部署正在HyperV中流量为双层隧道，内部流量为K8s虚拟机的存在相同IP，流聚合不能忽略MAC
    // 腾讯TCE：
    //   GRE隧道流量中的mac地址为伪造，流聚合忽略MAC地址
    // IPIP隧道：
    //   在IPIP隧道封装场景下，外层MAC在腾讯TCE环境中存在不对称情况
    //   实际上IPIP没有隧道ID，因此可以肯定不存在IP冲突，忽略MAC也是合理的
    fn mac_match(
        meta_packet: &MetaPacket,
        config_ignore: (bool, bool),
        trident_type: TridentType,
    ) -> MatchMac {
        let ignore_mac = meta_packet.tunnel.is_some()
            && ((Self::is_hyper_v(trident_type) && meta_packet.tunnel.unwrap().tier < 2)
                || meta_packet.tunnel.unwrap().tunnel_type == TunnelType::TencentGre
                || meta_packet.tunnel.unwrap().tunnel_type == TunnelType::Ipip);

        // return value stands different match type, defined by MAC_MATCH_*
        // TODO: maybe should consider L2End0 and L2End1 when InPort == 0x30000
        let is_from_isp = meta_packet.lookup_key.tap_type != TapType::Tor;
        if is_from_isp || ignore_mac || config_ignore.1 {
            return MatchMac::None;
        }

        let is_from_trident = meta_packet.lookup_key.tap_type == TapType::Tor
            && meta_packet.tap_port.split_fields().0 > 0;

        if !config_ignore.0 && is_from_trident {
            if !meta_packet.lookup_key.l2_end_0 && !meta_packet.lookup_key.l2_end_1 {
                return MatchMac::None;
            } else if !meta_packet.lookup_key.l2_end_0 {
                return MatchMac::Dst;
            } else {
                return MatchMac::Src;
            }
        }
        MatchMac::All
    }

    fn mac_match_with_direction(
        meta_packet: &MetaPacket,
        flow_mac_src: MacAddr,
        flow_mac_dst: MacAddr,
        match_mac: MatchMac,
        direction: PacketDirection,
    ) -> bool {
        let (src_mac, dst_mac) = match direction {
            PacketDirection::ClientToServer => (flow_mac_src, flow_mac_dst),
            PacketDirection::ServerToClient => (flow_mac_dst, flow_mac_src),
        };

        match match_mac {
            MatchMac::Dst => dst_mac == meta_packet.lookup_key.dst_mac,
            MatchMac::Src => src_mac == meta_packet.lookup_key.src_mac,
            MatchMac::All => {
                dst_mac == meta_packet.lookup_key.dst_mac
                    && src_mac == meta_packet.lookup_key.src_mac
            }
            MatchMac::None => true,
        }
    }

    fn endpoint_match_with_direction(
        peers: &[FlowMetricsPeer; 2],
        meta_packet: &MetaPacket,
        direction: PacketDirection,
    ) -> bool {
        if meta_packet.tunnel.is_none() {
            return true;
        }

        // 同一个TapPort上的流量，如果有隧道的话，当Port做发卡弯转发时，进出的内层流量完全一样
        // 此时需要额外比较L2End确定哪股是进入的哪股是出去的
        let lookup_key = &meta_packet.lookup_key;
        match direction {
            PacketDirection::ClientToServer => {
                lookup_key.l2_end_0 == peers[0].is_l2_end
                    && lookup_key.l2_end_1 == peers[1].is_l2_end
            }
            PacketDirection::ServerToClient => {
                lookup_key.l2_end_0 == peers[1].is_l2_end
                    && lookup_key.l2_end_1 == peers[0].is_l2_end
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use std::net::{Ipv4Addr, Ipv6Addr};
    use std::str::FromStr;

    use super::*;

    // tap_type = TapType::ISP(7), tap_port = 2100, src_mac = B0-60-88-51-D7-54 dst_mac = 00-15-5D-70-01-03
    // src_ipv4addr = 192.168.66.1 dst_ipv4addr = 192.168.66.2 src_port = 19001, dst_port = 19002
    // src_ipv6addr =  fe80::88d3:f197:5843:f873 dst_ipv6addr = fe80::742a:d20d:8d45:56e6
    fn new_map_key(eth_type: EthernetType, src_addr: IpAddr, dst_addr: IpAddr) -> FlowMapKey {
        let lookup_key = LookupKey {
            tap_type: TapType::Isp(7),
            src_mac: MacAddr::from([0xb0, 0x60, 0x88, 0x51, 0xd7, 0x54]),
            dst_mac: MacAddr::from([0x00, 0x15, 0x5d, 0x70, 0x01, 0x03]),
            src_ip: src_addr,
            dst_ip: dst_addr,
            src_port: 19001,
            dst_port: 19002,
            eth_type,
            ..Default::default()
        };
        FlowMapKey::new(&lookup_key, TapPort(2100))
    }

    #[test]
    fn ipv4_node_hash() {
        let key = new_map_key(
            EthernetType::Ipv4,
            Ipv4Addr::new(192, 168, 66, 1).into(),
            Ipv4Addr::new(192, 168, 66, 2).into(),
        );
        // 右边是go 版本计算得出
        assert_eq!(key.0, 0xecb912dddb15b140);
    }

    #[test]
    fn ipv6_node_hash() {
        let key = new_map_key(
            EthernetType::Ipv6,
            Ipv6Addr::from_str("fe80::88d3:f197:5843:f873")
                .unwrap()
                .into(),
            Ipv6Addr::from_str("fe80::742a:d20d:8d45:56e6")
                .unwrap()
                .into(),
        );
        // 右边是go 版本计算得出
        assert_eq!(key.0, 0xe7f0aea2897fd9ad);
    }

    #[test]
    fn arp_node_hash() {
        let key = new_map_key(
            EthernetType::Arp,
            Ipv6Addr::from_str("fe80::88d3:f197:5843:f873")
                .unwrap()
                .into(),
            Ipv6Addr::from_str("fe80::742a:d20d:8d45:56e6")
                .unwrap()
                .into(),
        );
        // 右边是go 版本计算得出
        assert_eq!(key.0, 1098954493523811076);
    }

    #[test]
    fn other_node_hash() {
        let key = new_map_key(
            EthernetType::Dot1Q,
            Ipv6Addr::from_str("fe80::88d3:f197:5843:f873")
                .unwrap()
                .into(),
            Ipv6Addr::from_str("fe80::742a:d20d:8d45:56e6")
                .unwrap()
                .into(),
        );
        // 右边是go 版本计算得出
        assert_eq!(key.0, 4948968142922745785);
    }
}
