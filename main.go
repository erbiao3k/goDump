package main

import (
	"context"
	"log"
	"time"

	"github.com/allegro/bigcache/v3"
	"goDump/config"
	k "goDump/kubernetes"
	"goDump/public"
	"goDump/web/router"
	v1 "k8s.io/api/core/v1"
)

func main() {
	go router.Router()

	cache, err := bigcache.New(context.Background(), bigcache.Config{
		Shards:      2,
		LifeWindow:  24 * time.Hour,
		CleanWindow: 3 * time.Second,
	})
	if err != nil {
		panic("初始化内存缓存失败：" + err.Error())
	}

	metricClient, err := k.NewMetricClient(k.RestConfig())
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}
	defer metricClient.Close()
	kubeClient, err := k.NewKubeClient(k.RestConfig())
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}
	defer kubeClient.Close()

	for {
		var problemPodList []k.PodResourceInfo

		podMap := make(map[string]v1.Pod)

		for _, ns := range config.Namespace {
			podListResponse, err := k.GetPodList(kubeClient.Client, ns)
			if err != nil {
				log.Println("Failed to get pod list for namespace ", ns, ": ", err)
				continue
			}
			podMetricsResponse, err := k.GetPodMetricsList(metricClient.Client, ns)
			if err != nil {
				log.Println("Failed to list pod metrics for namespace ", ns, ": ", err)
				continue
			}

			for _, pod := range podListResponse.Items {
				if !public.IgnorePod(pod.Name) {
					pod.Namespace = ns
					podMap[pod.Name] = pod
				}
			}

			for _, pm := range podMetricsResponse.Items {
				pm.Namespace = ns
				for _, container := range podMap[pm.Name].Spec.Containers {
					containerMetric := public.FindContainerMetric(pm.Containers, container.Name)
					if containerMetric != nil {
						public.ProcessContainer(cache, podMap[pm.Name], container, containerMetric, &problemPodList)
					}
				}
			}
		}

		public.WebCheckPodList(&problemPodList)

		public.HandleProblemPods(kubeClient, problemPodList, cache)

		time.Sleep(1 * time.Minute)
	}
}
