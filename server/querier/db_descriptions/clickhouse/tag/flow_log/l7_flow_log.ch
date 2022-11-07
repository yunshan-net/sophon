# Name                    , DisplayName              , Discription
_id                       ,                          ,
time                      ,                          ,

region                    , 区域                     ,
az                        , 可用区                   ,
host                      , 宿主机                   ,
chost                     , 云服务器                 ,
vpc                       , VPC                      ,
l2_vpc                    , 网卡VPC                  ,
subnet                    , 子网                     ,
router                    , 路由器                   ,
dhcpgw                    , DHCP网关                 ,
lb                        , 负载均衡器               ,
lb_listener               , 负载均衡监听器           ,
natgw                     , NAT网关                  ,
redis                     , Redis                    ,
rds                       , RDS                      ,
pod_cluster               , K8s容器集群              ,
pod_ns                    , K8s命名空间              ,
pod_node                  , K8s容器节点              ,
pod_ingress               , K8s Ingress              ,
pod_service               , K8s容器服务              ,
pod_group                 , K8s工作负载              ,
pod                       , K8s容器POD               ,
service                   , 服务                     ,
resource_gl0_type         , 类型-容器POD优先         ,
resource_gl0              , 资源-容器POD优先         ,
resource_gl1_type         , 类型-工作负载优先        ,
resource_gl1              , 资源-工作负载优先        ,
resource_gl2_type         , 类型-服务优先            ,
resource_gl2              , 资源-服务优先            ,
process_id                , 进程ID                   ,
process_kname             , 线程名                   ,
service_name              , 服务名称                 ,
service_instance_id       , 服务实例                 ,
is_internet               , Internet标志             ,

ip                        , IP地址                   ,
tunnel_type               , 隧道类型                 ,
protocol                  , 网络协议                 ,

client_port               , 客户端口                 ,
server_port               , 服务端口                 ,
req_tcp_seq               , 请求TCP包SEQ             ,
resp_tcp_seq              , 响应TCP包SEQ             ,

l7_protocol               , 应用协议                 ,
l7_protocol_str           , 应用协议                 ,
version                   , 协议版本                 ,
type                      , 日志类型                 ,
request_type              , 请求类型                 ,
request_domain            , 请求域名                 ,
request_resource          , 请求资源                 ,
endpoint                  , 端点                     ,
request_id                , 请求ID                   ,
response_status           , 响应状态                 ,
response_code             , 响应码                   ,
response_exception        , 响应异常                 ,
response_result           , 响应结果                 ,

http_proxy_client         , HTTP代理客户端           ,
trace_id                  , TraceID                  ,
span_id                   , SpanID                   ,
parent_span_id            , ParentSpanID             ,
span_kind                 , Span类型                 ,

x_request_id              , XRequestID               ,
syscall_trace_id_request  , SyscallTraceID-请求      ,
syscall_trace_id_response , SyscallTraceID-响应      ,
syscall_thread_0          , Syscall线程-请求         ,
syscall_thread_1          , Syscall线程-响应         ,
syscall_cap_seq_0         , Syscall序列号-请求       ,
syscall_cap_seq_1         , Syscall序列号-响应       ,

flow_id                   , 流日志ID                 ,
tap                       , 采集点                   ,

tap_port_type             , 采集位置类型             ,
tap_port_name             , 采集位置名称             ,
tap_port                  , 采集位置标识             ,
vtap                      , 采集器                   ,
start_time                , 开始时间                 ,
end_time                  , 结束时间                 ,
tap_side                  , 路径统计位置             ,
attributes                , Attributes               ,
labels                    , K8s Labels               ,
