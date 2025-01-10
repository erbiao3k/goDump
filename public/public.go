package public

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/allegro/bigcache/v3"
	"goDump/config"
	"goDump/kubernetes"
	"k8s.io/api/core/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func genMd5(str string) string {
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}

func IgnorePod(podStr string) bool {
	prefixs := []string{"xod-es-", "xod-nacos-", "xod-rabbitmq-", "xod-skywalking-", "aly-billing-monitor-", "aly-ssl-monitor-", "xod-export-mq-consum", "ec-mysqldata-backup-", "godump-", "ops-alert-center-", "xod-api-", "xod-mindoc-", "xod-oms-ui", "xod-wms-ui"}
	for _, p := range prefixs {
		if strings.HasPrefix(podStr, p) {
			return true
		}
	}
	return false
}

func FindContainerMetric(containerMetrics []v1beta1.ContainerMetrics, name string) *v1beta1.ContainerMetrics {
	for _, cm := range containerMetrics {
		if cm.Name == name {
			return &cm
		}
	}
	return nil
}

func ProcessContainer(cache *bigcache.BigCache, pod v1.Pod, container v1.Container, containerMetric *v1beta1.ContainerMetrics, problemPodList *[]kubernetes.PodResourceInfo) {
	podResourceInfo := kubernetes.PodResourceInfo{
		PodName:       pod.Name,
		ContainerName: container.Name,
		Namespace:     pod.Namespace,
		PodIp:         pod.Status.PodIP,
		HostIp:        pod.Status.HostIP,
		CPURequest:    container.Resources.Requests.Cpu(),
		CPULimit:      container.Resources.Limits.Cpu(),
		CPUUsage:      containerMetric.Usage.Cpu(),
		MemoryRequest: container.Resources.Requests.Memory(),
		MemoryLimit:   container.Resources.Limits.Memory(),
		MemoryUsage:   containerMetric.Usage.Memory(),
	}

	md5Value := genMd5(podResourceInfo.Namespace + podResourceInfo.PodName + podResourceInfo.ContainerName)

	cached, err := cache.Get(md5Value)
	if err != nil || len(cached) == 0 {
		if podResourceInfo.MemoryUsagePercent > 0.85 {
			podResourceInfo.ProblemItem = "memory"
			podResourceInfo.Command = fmt.Sprintf(config.MemoryCpuCmd, config.CurrentPodIp)
			*problemPodList = append(*problemPodList, podResourceInfo)
			cache.Set(md5Value, []byte("1"))
		}
		if podResourceInfo.CPUUsagePercent > 0.85 {
			podResourceInfo.ProblemItem = "cpu"
			podResourceInfo.Command = fmt.Sprintf(config.CpuCmd, config.CurrentPodIp)
			*problemPodList = append(*problemPodList, podResourceInfo)
			cache.Set(md5Value, []byte("1"))
		}
	}
}

func WebCheckPodList(problemPodList *[]kubernetes.PodResourceInfo) {
	for _, pod := range config.CheckPodList {
		*problemPodList = append(*problemPodList, kubernetes.PodResourceInfo{
			ProblemItem:   "noProblemFromWeb",
			PodName:       pod.Pod,
			ContainerName: pod.Container,
			Namespace:     pod.Namespace,
			Command:       fmt.Sprintf(config.MemoryCpuCmd, config.CurrentPodIp),
		})
	}
	config.CheckPodList = []config.CheckPod{}
}

func HandleProblemPods(kubeClient *kubernetes.KubeClient, problemPodList []kubernetes.PodResourceInfo, cache *bigcache.BigCache) {
	if len(problemPodList) > 0 {
		log.Println("need goDump pod:  ", problemPodList)
		for _, pod := range problemPodList {
			md5Value := genMd5(pod.Namespace + pod.PodName + pod.ContainerName + pod.ProblemItem)
			cached, err := cache.Get(md5Value)
			if err != nil || len(cached) == 0 || string(cached) == "fail" {
				opts := &kubernetes.ExecOptions{
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
				stdout, stderr, err := kubernetes.ExecInPod(kubeClient, opts, kubernetes.RestConfig())
				if err != nil {
					fmt.Printf("Error executing command: %v\n", err)
					cache.Set(md5Value, []byte("fail"))
				} else {
					fmt.Printf("Command output: %s\n", stdout)
					fmt.Printf("Command error: %s\n", stderr)
					err = send(pod.Namespace, pod.PodName, pod.PodIp, pod.ProblemItem,
						config.DownloadHost+"/download/"+pod.PodName+"-jvm-dump.tar.gz",
						config.SlsAddr,
						config.MonitorAddr,
						config.K8sAddr,
						"")
					if err != nil {
						log.Println("群机器人消息推送失败，err：", err)
					}
					cache.Set(md5Value, []byte("success"))
				}
			}
		}
	}
}

func send(ns, pod, podIp, item, dumpAddr, slsAddr, monitorAddr, k8sAddr, people string) error {
	// 设置随机种子
	rand.Seed(time.Now().UnixNano())
	// 产生一个 [0, len(wecomRobotAddr)-1) 的随机整数
	randNum := rand.Intn(len(config.WecomRobotAddr) - 1)

	postData := fmt.Sprintf(`{"msgtype": "text", "text": {"content": "【Java应用告警】\n命名空间：%s\nPOD名称：%s\nPODIP：%s\n异常指标：%s\nDUMP数据：%s\nSLS平台：%s\n监控平台：%s\nK8S平台：%s","mentioned_mobile_list":["%s"]}}`, ns, pod, podIp, item, dumpAddr, slsAddr, monitorAddr, k8sAddr, people)

	if _, err := http.Post(config.WecomRobotAddr[randNum], "application/json", strings.NewReader(postData)); err != nil {
		return err
	}
	return nil
}

//	postData := fmt.Sprintf(`{"msgtype": "markdown", "markdown": {"content": "<font color=\"warning\">Java应用告警</font>\n> 命名空间：<font color=\"comment\">%s</font>\n> POD名称：<font color=\"comment\">%s</font>\n> 异常指标：<font color=\"comment\">%s</font>\n> DUMP数据：<font color=\"comment\">[下载](%s)</font>\n> SLS平台：<font color=\"comment\">[跳转](%s)</font>\n> 监控平台：<font color=\"comment\">[跳转](%s)</font>\n> K8S平台：<font color=\"comment\">[跳转](%s)</font>","mentioned_mobile_list":["%s"]}}`, ns, pod, item, dumpAddr, slsAddr, monitorAddr, k8sAddr, people)
