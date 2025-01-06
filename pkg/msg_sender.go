package pkg

import (
	"fmt"
	"goDump/config"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func Send(ns, pod, podIp, item, dumpAddr, slsAddr, monitorAddr, k8sAddr, people string) error {
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
