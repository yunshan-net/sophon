# Name                , DisplayName                  , Category               , Description
_id                   ,                              ,                        ,
time                  ,                              ,                        ,
       
region                , 区域                          , 知识图谱               ,
az                    , 可用区                        , 知识图谱               ,
host                  , 宿主机                        , 知识图谱               ,
chost                 , 云服务器                      , 知识图谱               ,
vpc                   , VPC                          , 知识图谱               ,
l2_vpc                , 网卡VPC                      , 知识图谱                ,
subnet                , 子网                         , 知识图谱                ,
router                , 路由器                       , 知识图谱                ,
dhcpgw                , DHCP网关                     , 知识图谱                ,
lb                    , 负载均衡器                   , 知识图谱                ,
lb_listener           , 负载均衡监听器                , 知识图谱               ,
natgw                 , NAT网关                      , 知识图谱               ,
redis                 , Redis                        , 知识图谱               ,
rds                   , RDS                          , 知识图谱               ,
pod_cluster           , K8s容器集群                  , 知识图谱               ,
pod_ns                , K8s命名空间                  , 知识图谱               ,
pod_node              , K8s容器节点                  , 知识图谱               ,
pod_ingress           , K8s Ingress                  , 知识图谱               ,
pod_service           , K8s容器服务                  , 知识图谱               ,
pod_group             , K8s工作负载                  , 知识图谱               ,
pod                   , K8s容器POD                   , 知识图谱               ,
service               , 服务                         , 知识图谱               ,
resource_gl0_type     , 类型-容器POD优先              , 知识图谱                ,
resource_gl0          , 资源-容器POD优先             , 知识图谱                ,
resource_gl1_type     , 类型-工作负载优先            , 知识图谱                 , 
resource_gl1          , 资源-工作负载优先            , 知识图谱                 , 
resource_gl2_type     , 类型-服务优先                , 知识图谱                 ,    
resource_gl2          , 资源-服务优先                , 知识图谱                 ,
is_internet           , Internet标志                 ,                        ,
       
mac                   , MAC                          , 链路层                 ,
eth_type              , 链路协议                      , 链路层                 ,
vlan                  , VLAN                         , 链路层                 ,
       
ip                    , IP地址                       , 网络层                 ,
is_ipv4               , IP类型                       , 网络层                 ,
protocol              , 网络协议                     , 网络层                 ,
tunnel_tier           , 隧道层数                     , 网络层                 ,
tunnel_type           , 隧道类型                     , 网络层                 ,
tunnel_tx_id          , 请求隧道ID                   , 网络层                 ,
tunnel_rx_id          , 响应隧道ID                   , 网络层                 ,
tunnel_tx_ip_0        , 请求隧道源IP                 , 网络层                 ,
tunnel_tx_ip_1        , 请求隧道目IP                 , 网络层                 ,
tunnel_rx_ip_0        , 响应隧道源IP                 , 网络层                 ,
tunnel_rx_ip_1        , 响应隧道目IP                 , 网络层                 ,
tunnel_tx_mac_0       , 请求隧道源MAC                , 网络层                 ,
tunnel_tx_mac_1       , 请求隧道目MAC                , 网络层                 ,
tunnel_rx_mac_0       , 响应隧道源MAC                , 网络层                 ,
tunnel_rx_mac_1       , 响应隧道目MAC                , 网络层                 ,
       
       
client_port           , 客户端口                     , 传输层                 ,
server_port           , 服务端口                     , 传输层                 ,
tcp_flags_bit         , TCP标志位列表                , 传输层                 ,
syn_seq               , SYN包SEQ                     , 传输层                 ,
syn_ack_seq           , SYN-ACK包SEQ                 , 传输层                 ,
last_keepalive_seq    , 心跳SEQ                      , 传输层                 ,
last_keepalive_ack    , 心跳ACK                      , 传输层                 ,
       
l7_protocol           , 应用协议                     , 应用层                 ,
       
province              , 省份                         , 广域网                 ,
       
close_type            , 流结束类型                   , 流信息                  ,
flow_source           , 流数据来源                   , 流信息                 ,
flow_id               , 流日志ID                     , 流信息                 ,
tap                   , 采集点                       , 流信息                 ,
tap_port              , 采集位置标识                 , 流信息                 ,
tap_port_name         , 采集位置名称                 , 流信息                 ,
tap_port_type         , 采集位置类型                 , 流信息                 ,
vtap                  , 采集器                       , 流信息                 ,
       
tap_side              , 路径统计位置                 , 流信息                 ,
l2_end                , 二层边界                     , 流信息                 ,
l3_end                , 三层边界                     , 流信息                 ,
is_new_flow           , 新建流                       , 流信息                 ,
start_time            , 开始时间                     , 流信息                 ,
end_time              , 结束时间                     , 流信息                 ,
duration              , 流持续时间                   , 流信息                 , 单位: 微秒
status                , 状态                         , 流信息                 , 由流结束类型决定。正常：正常结束、周期性上报。客户端异常：客户端SYN结束、客户端重置、客户端半关、客户端端口复用、客户端其他重置。服务端异常：服务端重置、连接超时、服务端半关、服务端SYN结束、服务端直接重置、服务端队列溢出、服务端其他重置。未知：其他结束方式。
labels                , K8s Labels                   , K8s Labels            , 
