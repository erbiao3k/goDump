package config

import (
	"context"
	"github.com/allegro/bigcache/v3"
	"os"
	"strings"
	"time"
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

func init() {
	if _, err := os.Stat(ShareDir); os.IsNotExist(err) {
		err := os.MkdirAll(ShareDir, 0755)
		if err != nil {
			panic("Failed to create directory: " + err.Error())
		}
	}
}

var TimedCache = func() *bigcache.BigCache {
	cache, err := bigcache.New(context.Background(), bigcache.Config{Shards: 2, LifeWindow: 24 * time.Hour, CleanWindow: 3 * time.Second})
	if err != nil {
		panic("初始化定时任务内存缓存失败：" + err.Error())
	}
	return cache
}

var WebCache = func() *bigcache.BigCache {
	cache, err := bigcache.New(context.Background(), bigcache.Config{Shards: 2, LifeWindow: 10 * time.Minute, CleanWindow: 3 * time.Second})
	if err != nil {
		panic("初始化定时任务内存缓存失败：" + err.Error())
	}
	return cache
}
