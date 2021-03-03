package metrics

import (
	"io"

	core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/scheme"
)

type writeWrapper struct {
	data []byte
}

func (w *writeWrapper) Write(data []byte) (int, error) {
	w.data = append(w.data, data...)
	return len(data), nil
}

func newWriteWrapper() writeWrapper {
	return writeWrapper{make([]byte, 0)}
}

func execCmd(
	client kubernetes.Interface,
	config *restclient.Config,
	podName string,
	command string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
	containerName string,
) error {
	cmd := []string{
		"sh",
		"-c",
		command,
	}
	req := client.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
		Namespace("default").SubResource("exec")
	option := &core.PodExecOptions{
		Command:   cmd,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
		Container: containerName,
	}
	if stdin == nil {
		option.Stdin = false
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return err
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	})
	if err != nil {
		return err
	}

	return nil
}
