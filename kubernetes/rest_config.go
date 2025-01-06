package kubernetes

import (
	"fmt"
	"k8s.io/client-go/rest"
)

// ConfigHolder 持有 Kubernetes 配置对象
type ConfigHolder struct {
	Config *rest.Config
}

// GlobalConfig 单例模式的配置持有者
var GlobalConfig = ConfigHolder{}

// GetConfig 获取配置对象
func (ch *ConfigHolder) GetConfig() (*rest.Config, error) {
	if ch.Config != nil {
		return ch.Config, nil
	}
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("get config object from cluster failed: %w", err)
	}
	ch.Config = cfg
	return cfg, nil
}

var RestConfig = func() *rest.Config {
	cfg, err := GlobalConfig.GetConfig()
	if err != nil {
		panic(err)
	}
	return cfg
}
