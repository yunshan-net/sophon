use thiserror::Error;

#[derive(Debug, Error)]
pub enum Error {
    #[error("parse mac address failed from: {0}")]
    ParseMacFailed(String),
    #[error("call try_from() failed from {0}")]
    TryFromFailed(String),
    #[error(transparent)]
    KubeWatcher(#[from] kube::runtime::watcher::Error),
    #[error(transparent)]
    ParseUtf8(#[from] std::string::FromUtf8Error),
    #[error("PlatformSynchronizer failed: {0} ")]
    PlatformSynchronizer(String),
    #[error(transparent)]
    IoError(#[from] std::io::Error),
    #[error("data not found: {0}")]
    NotFound(String),
    #[error("Kubernetes ApiWatcher error: {0}")]
    KubernetesApiWatcher(String),
    #[error("system: {0}")]
    SysMonitor(String),
    #[error("environment error: {0}")]
    Environment(String),
    #[error(transparent)]
    Errno(#[from] nix::errno::Errno),
    #[error("ethtool: {0}")]
    Ethtool(String),
    #[error("parse packet failed from: {0}")]
    ParsePacketFailed(String),
    #[error("dns perf parse: {0}")]
    DnsPerfParse(String),
    #[error("dns log parse: {0}")]
    DnsLogParse(String),
    #[error("redis perf parse: {0}")]
    RedisPerfParse(String),
    #[error("redis log parse: {0}")]
    RedisLogParse(String),
    #[error("kafka perf parse: {0}")]
    KafkaPerfParse(String),
    #[error("no kafka log {0}")]
    KafkaLogParse(String),
    #[error("invalid tpacket version: {0}")]
    InvalidTpVersion(isize),
    #[error("dubbo parse: {0}")]
    DubboParse(String),
    #[error("dubbo perf parse: {0}")]
    DubboPerfParse(String),
    #[error("dubbo log parse: {0}")]
    DubboLogParse(String),
}

pub type Result<T, E = Error> = std::result::Result<T, E>;
