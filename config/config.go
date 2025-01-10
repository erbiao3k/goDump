package config

import (
	"os"
	"strings"
)

type CheckPod struct {
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Container string `json:"container"`
}

var (
	ShareDir       = "./goDumpShareDir/"
	CurrentPodIp   = os.Getenv("POD_IP")
	Namespace      = strings.Split(os.Getenv("NAMESPACE"), ",")
	DownloadHost   = os.Getenv("DOWNLOAD_HOST")
	WecomRobotAddr = strings.Split(os.Getenv("WECOM_ROBOT_ADDR"), ",")
	CheckPodList   []CheckPod
)

var (
	MemoryCpuCmd = `pidof java >> /dev/null && rm -rf /tmp/goDump && mkdir /tmp/goDump && curl -sS -O https://arthas.aliyun.com/arthas-boot.jar && java -jar arthas-boot.jar -c 'jvm > /tmp/goDump/jvm-info.log;thread -b > /tmp/goDump/blocking-thread.log;thread -n 5 > /tmp/goDump/busy-thread.log;heapdump /tmp/goDump/memory-dump.hprof;stop' -v $(pidof java) && tar -zcf /tmp/$HOSTNAME-jvm-dump.tar.gz /tmp/goDump/ && curl -X POST -F "upload[]=@/tmp/$HOSTNAME-jvm-dump.tar.gz" http://%s:12399/upload && rm -rf /tmp/$HOSTNAME-jvm-dump.tar.gz /tmp/goDump`
	CpuCmd       = `pidof java >> /dev/null && rm -rf /tmp/goDump && mkdir /tmp/goDump && curl -sS -O https://arthas.aliyun.com/arthas-boot.jar && java -jar arthas-boot.jar -c 'jvm > /tmp/goDump/jvm-info.log;thread -b > /tmp/goDump/blocking-thread.log;thread -n 5 > /tmp/goDump/busy-thread.log;stop' -v $(pidof java) && tar -zcf /tmp/$HOSTNAME-jvm-dump.tar.gz /tmp/goDump/ && curl -X POST -F "upload[]=@/tmp/$HOSTNAME-jvm-dump.tar.gz" http://%s:12399/upload && rm -rf /tmp/$HOSTNAME-jvm-dump.tar.gz /tmp/goDump`
)

const (
	SlsAddr     = "https://sls.console.aliyun.com/lognext/project/"
	MonitorAddr = "https://csnew.console.aliyun.com/#/next/clusters/"
	K8sAddr     = "https://csnew.console.aliyun.com/#/k8s/cluster/"
)

func init() {
	if _, err := os.Stat(ShareDir); os.IsNotExist(err) {
		err := os.MkdirAll(ShareDir, 0755)
		if err != nil {
			panic("Failed to create directory: " + err.Error())
		}
	}
}
