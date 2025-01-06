package kubernetes

import (
	"bytes"
	"io"
	PodExecOptionsv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// ExecOptions 包含执行 Pod 内部命令所需的选项
type ExecOptions struct {
	Command   []string
	Container string
	Stdin     bool
	Stdout    bool
	Stderr    bool
	TTY       bool
	PodName   string
	Namespace string
}

// ExecInPod 在指定的 Pod 内部执行命令
func ExecInPod(client *KubeClient, opts *ExecOptions, config *rest.Config) (string, string, error) {
	// 构建 exec 请求

	req := client.Client.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(opts.PodName).
		Namespace(opts.Namespace).
		SubResource("exec").
		VersionedParams(&PodExecOptionsv1.PodExecOptions{
			Command:   opts.Command,
			Container: opts.Container,
			Stdin:     opts.Stdin,
			Stdout:    opts.Stdout,
			Stderr:    opts.Stderr,
			TTY:       opts.TTY,
		}, scheme.ParameterCodec)

	// 创建执行器
	executor, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return "", "", err
	}

	// 执行命令并捕获输出
	var stdout, stderr bytes.Buffer
	err = executor.Stream(remotecommand.StreamOptions{
		Stdin:  io.Reader(nil),
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    opts.TTY,
	})

	// 返回命令输出和错误信息
	return stdout.String(), stderr.String(), err
}
