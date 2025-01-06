package kubernetes

import (
	"fmt"
	"k8s.io/client-go/rest"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

// MetricClient 用于与 Kubernetes Metrics API 交互
type MetricClient struct {
	Client *versioned.Clientset
}

// NewMetricClient 创建一个新的 MetricClient
func NewMetricClient(cfg *rest.Config) (*MetricClient, error) {
	client, err := versioned.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("creates metric client from the given config failed: %w", err)
	}
	return &MetricClient{
		Client: client,
	}, nil
}

// Close 关闭 MetricClient 的底层 HTTP 客户端
func (m *MetricClient) Close() {
}
