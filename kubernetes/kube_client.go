package kubernetes

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// KubeClient 用于与 Kubernetes Core API 交互
type KubeClient struct {
	Client *kubernetes.Clientset
}

// NewKubeClient 创建一个新的 KubeClient
func NewKubeClient(cfg *rest.Config) (*KubeClient, error) {
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}
	return &KubeClient{
		Client: clientset,
	}, nil
}

// Close 关闭 KubeClient 的底层 HTTP 客户端
func (k *KubeClient) Close() {
}
