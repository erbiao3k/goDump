package kubernetes

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

// GetPodMetricsList 获取指定命名空间下的所有 Pod 指标列表
func GetPodMetricsList(client *versioned.Clientset, namespace string) (*v1beta1.PodMetricsList, error) {
	podMetricsList, err := client.MetricsV1beta1().PodMetricses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pod metrics: %w", err)
	}
	return podMetricsList, nil
}
