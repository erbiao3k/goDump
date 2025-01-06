package main

import (
	"fmt"
	"goDump/config"
	k "goDump/kubernetes"
	"goDump/pkg"
	"goDump/web/router"
	"log"
	"strings"
	"time"
)

func main() {
	go router.Router()
	for {
		metricClient, err := k.NewMetricClient(k.RestConfig())
		if err != nil {
			panic(err)
		}
		defer metricClient.Close()
		kubeClient, err := k.NewKubeClient(k.RestConfig())
		if err != nil {
			panic(err)
		}
		defer kubeClient.Close()

		var md5Value string
		var problemPodList []k.PodResourceInfo

		for _, ns := range config.Namespace {
			podList, err := k.GetPodList(kubeClient.Client, ns)
			if err != nil {
				panic(err)
			}

			podMetricsList, err := k.GetPodMetricsList(metricClient.Client, ns)
			if err != nil {
				panic(err)
			}

			for _, pod := range podList.Items {
				if isIgnorePod(pod.Name) {
					continue
				}
				for _, container := range pod.Spec.Containers {
					for _, pm := range podMetricsList.Items {
						if pm.Name == pod.Name {
							for _, containerMetric := range pm.Containers {
								if containerMetric.Name == container.Name {
									podResourceInfo := k.PodResourceInfo{
										PodName:            pod.Name,
										ContainerName:      container.Name,
										Namespace:          pod.Namespace,
										PodIp:              pod.Status.PodIP,
										HostIp:             pod.Status.HostIP,
										CPURequest:         container.Resources.Requests.Cpu(),
										CPULimit:           container.Resources.Limits.Cpu(),
										CPUUsage:           containerMetric.Usage.Cpu(),
										CPUUsagePercent:    k.CalculateResourceUsage(containerMetric.Usage.Cpu(), container.Resources.Limits.Cpu()),
										MemoryRequest:      container.Resources.Requests.Memory(),
										MemoryLimit:        container.Resources.Limits.Memory(),
										MemoryUsage:        containerMetric.Usage.Memory(),
										MemoryUsagePercent: k.CalculateResourceUsage(containerMetric.Usage.Memory(), container.Resources.Limits.Memory()),
									}
									md5Value = pkg.GenMd5(podResourceInfo.Namespace + podResourceInfo.PodName + podResourceInfo.ContainerName + podResourceInfo.ProblemItem)
									// 存在cache中则不添加到problemPodList
									cached, err := config.TimedCache().Get(md5Value)
									if err != nil || len(cached) == 0 {
										// 缓存中没有，需要处理
										if podResourceInfo.MemoryUsagePercent > 0.85 {
											podResourceInfo.ProblemItem = "memory"
											podResourceInfo.Command = fmt.Sprintf(`pidof java >> /dev/null && rm -rf /tmp/goDump && mkdir /tmp/goDump && curl -sS -O https://arthas.aliyun.com/arthas-boot.jar && java -jar arthas-boot.jar -c 'jvm > /tmp/goDump/jvm-info.log;thread -b > /tmp/goDump/blocking-thread.log;thread -n 5 > /tmp/goDump/busy-thread.log;heapdump /tmp/goDump/memory-dump.hprof;stop' -v $(pidof java) && tar -zcf /tmp/$HOSTNAME-jvm-dump.tar.gz /tmp/goDump/ && curl -X POST -F "upload[]=@/tmp/$HOSTNAME-jvm-dump.tar.gz" http://%s:12399/upload && rm -rf /tmp/$HOSTNAME-jvm-dump.tar.gz /tmp/goDump`, config.CurrentPodIp)
											problemPodList = append(problemPodList, podResourceInfo)
										}
										if podResourceInfo.CPUUsagePercent > 0.85 {
											podResourceInfo.ProblemItem = "cpu"
											podResourceInfo.Command = fmt.Sprintf(`pidof java >> /dev/null && rm -rf /tmp/goDump && mkdir /tmp/goDump && curl -sS -O https://arthas.aliyun.com/arthas-boot.jar && java -jar arthas-boot.jar -c 'jvm > /tmp/goDump/jvm-info.log;thread -b > /tmp/goDump/blocking-thread.log;thread -n 5 > /tmp/goDump/busy-thread.log;stop' -v $(pidof java) && tar -zcf /tmp/$HOSTNAME-jvm-dump.tar.gz /tmp/goDump/ && curl -X POST -F "upload[]=@/tmp/$HOSTNAME-jvm-dump.tar.gz" http://%s:12399/upload && rm -rf /tmp/$HOSTNAME-jvm-dump.tar.gz /tmp/goDump`, config.CurrentPodIp)
											problemPodList = append(problemPodList, podResourceInfo)
										}
									} else {
										// 缓存中已存在，跳过处理
										continue
									}
								}
							}
						}
					}
				}
			}
		}

		for _, pod := range config.CheckPodList {
			problemPodList = append(problemPodList, k.PodResourceInfo{
				ProblemItem:   "noProbelmFromWeb",
				PodName:       pod.Pod,
				ContainerName: pod.Container,
				Namespace:     pod.Namespace,
				Command:       fmt.Sprintf(`pidof java >> /dev/null && rm -rf /tmp/goDump && mkdir /tmp/goDump && curl -sS -O https://arthas.aliyun.com/arthas-boot.jar && java -jar arthas-boot.jar -c 'jvm > /tmp/goDump/jvm-info.log;thread -b > /tmp/goDump/blocking-thread.log;thread -n 5 > /tmp/goDump/busy-thread.log;heapdump /tmp/goDump/memory-dump.hprof;stop' -v $(pidof java) && tar -zcf /tmp/$HOSTNAME-jvm-dump.tar.gz /tmp/goDump/ && curl -X POST -F "upload[]=@/tmp/$HOSTNAME-jvm-dump.tar.gz" http://%s:12399/upload && rm -rf /tmp/$HOSTNAME-jvm-dump.tar.gz /tmp/goDump`, config.CurrentPodIp),
			})
		}
		config.CheckPodList = []config.CheckPod{}

		if len(problemPodList) > 0 {
			log.Println("need goDump pod:  ", problemPodList)
			for _, pod := range problemPodList {
				opts := &k.ExecOptions{
					Command:   []string{"sh", "-c", pod.Command},
					Container: pod.ContainerName,
					Stdin:     false,
					Stdout:    true,
					Stderr:    true,
					TTY:       false,
					PodName:   pod.PodName,
					Namespace: pod.Namespace,
				}
				log.Println("problem pod info: ", pod.Namespace, pod.PodName, pod.ContainerName, pod.ProblemItem, opts.Command)
				stdout, stderr, err := k.ExecInPod(kubeClient, opts, k.RestConfig())
				if err != nil {
					fmt.Printf("Error executing command: %v\n", err)
				} else {
					fmt.Printf("Command output: %s\n", stdout)
					fmt.Printf("Command error: %s\n", stderr)
					err = pkg.Send(pod.Namespace, pod.PodName, pod.PodIp, pod.ProblemItem,
						config.DownloadHost+"/download/"+pod.PodName+"-jvm-dump.tar.gz",
						"https://sls.console.aliyun.com/lognext/project/k8s-log-c22c0bb3b538441bdb62821a7dd8445e9/overview?slsRegion=cn-hongkong",
						"https://csnew.console.aliyun.com/#/next/clusters/c22c0bb3b538441bdb62821a7dd8445e9/monitoring/k8s-deploy", "https://csnew.console.aliyun.com/#/k8s/cluster/c22c0bb3b538441bdb62821a7dd8445e9/v2/workload/deployment/list?type=deployment&ns=appna",
						"")
					if err != nil {
						log.Println("群机器人消息推送失败，err：", err)
					}
					config.TimedCache().Set(md5Value, []byte("1"))
				}
			}
		}
		time.Sleep(1 * time.Minute)
	}
}

func isIgnorePod(podStr string) bool {
	prefixs := []string{"appna-es-", "appna-nacos-"}
	for _, p := range prefixs {
		if strings.HasPrefix(podStr, p) {
			return true
		}
	}
	return false
}
