package tag

var TAG_RESOURCE_TYPE_DEFAULT = []string{
	"region", "az", "pod_node", "pod_ns",
	"pod_group", "pod", "pod_cluster", "subnet",
}
var TAG_RESOURCE_TYPE_AUTO = []string{"resource_gl0", "resource_gl1", "resource_gl2"}

const (
	VIF_DEVICE_TYPE_INTERNET       = 0
	VIF_DEVICE_TYPE_VM             = 1
	VIF_DEVICE_TYPE_VROUTER        = 5
	VIF_DEVICE_TYPE_HOST           = 6
	VIF_DEVICE_TYPE_DHCP_PORT      = 9
	VIF_DEVICE_TYPE_POD            = 10
	VIF_DEVICE_TYPE_POD_SERVICE    = 11
	VIF_DEVICE_TYPE_REDIS_INSTANCE = 12
	VIF_DEVICE_TYPE_RDS_INSTANCE   = 13
	VIF_DEVICE_TYPE_POD_NODE       = 14
	VIF_DEVICE_TYPE_LB             = 15
	VIF_DEVICE_TYPE_NAT_GATEWAY    = 16
	VIF_DEVICE_TYPE_POD_GROUP      = 101
	VIF_DEVICE_TYPE_SERVICE        = 102
	VIF_DEVICE_TYPE_IP             = 255
)

var AutoMap = map[string]int{
	"vm":          VIF_DEVICE_TYPE_VM,
	"router":      VIF_DEVICE_TYPE_VROUTER,
	"host":        VIF_DEVICE_TYPE_HOST,
	"dhcp_port":   VIF_DEVICE_TYPE_DHCP_PORT,
	"pod_service": VIF_DEVICE_TYPE_POD_SERVICE,
	"redis":       VIF_DEVICE_TYPE_REDIS_INSTANCE,
	"rds":         VIF_DEVICE_TYPE_RDS_INSTANCE,
	"pod_node":    VIF_DEVICE_TYPE_POD_NODE,
	"lb":          VIF_DEVICE_TYPE_LB,
	"nat_gateway": VIF_DEVICE_TYPE_NAT_GATEWAY,
	"internet":    VIF_DEVICE_TYPE_INTERNET,
}

var AutoPodMap = map[string]int{
	"pod": VIF_DEVICE_TYPE_POD,
}

var AutoPodGroupMap = map[string]int{
	"pod_group": VIF_DEVICE_TYPE_POD_GROUP,
}

var AutoServiceMap = map[string]int{
	"pod_group": VIF_DEVICE_TYPE_POD_GROUP,
	"service":   VIF_DEVICE_TYPE_SERVICE,
}
