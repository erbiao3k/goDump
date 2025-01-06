package kubernetes

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

type PodResourceInfo struct {
	// cpu,memory
	ProblemItem        string             `json:"problem_item"`
	Command            string             `json:"command,omitempty"`
	PodName            string             `json:"pod_name"`
	ContainerName      string             `json:"container_name"`
	Namespace          string             `json:"namespace"`
	PodIp              string             `json:"pod_ip"`
	HostIp             string             `json:"host_ip"`
	CPURequest         *resource.Quantity `json:"cpu_request"`
	CPULimit           *resource.Quantity `json:"cpu_limit"`
	CPUUsage           *resource.Quantity `json:"cpu_usage"`
	CPUUsagePercent    float64            `json:"cpu_usage_percent"`
	MemoryRequest      *resource.Quantity `json:"memory_request"`
	MemoryLimit        *resource.Quantity `json:"memory_limit"`
	MemoryUsage        *resource.Quantity `json:"memory_usage"`
	MemoryUsagePercent float64            `json:"memory_usage_percent"`
}

// CalculateResourceUsage 计算资源使用百分比
func CalculateResourceUsage(used, limit *resource.Quantity) float64 {
	return float64(used.MilliValue()) / float64(limit.MilliValue())
}
