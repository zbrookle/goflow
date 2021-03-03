package metrics

import (
	"context"
	// "fmt"
	"io"

	// "net/url"

	"os"

	core "k8s.io/api/core/v1"
	k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
)

// DAGMetricsClient handles all interactions with DAG metrics
type DAGMetricsClient struct {
	kubeClient kubernetes.Interface
}

// PodMetrics holds information about the resource usage of a given pod
type PodMetrics struct {
	Memory int
	CPU    float32
}

// RemoteExecutor is an executor for running commands on containers
type RemoteExecutor struct{}

func execCmd(client kubernetes.Interface, config *restclient.Config, podName string,
	command string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	cmd := []string{
		"sh",
		"-c",
		command,
	}
	req := client.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
		Namespace("default").SubResource("exec")
	option := &core.PodExecOptions{
		Command: cmd,
		Stdin:   true,
		Stdout:  true,
		Stderr:  true,
		TTY:     true,
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

// NewDAGMetricsClient returns a new DAGMetricsClient from a metrics clientset
func NewDAGMetricsClient(clientSet kubernetes.Interface) *DAGMetricsClient {
	return &DAGMetricsClient{clientSet}
}

func getRestConfig() *restclient.Config {
	kubeConfigFlags := genericclioptions.NewConfigFlags(
		true,
	).WithDeprecatedPasswordFlag() // TODO: Figure out how to set up config without using kubectl
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	commandFactory := cmdutil.NewFactory(matchVersionKubeConfigFlags)
	restConfig, err := commandFactory.ToRESTConfig()
	if err != nil {
		panic(err)
	}
	return restConfig
}

// ListPodMetrics returns a list of all metrics for pods in a given namespace
func (client *DAGMetricsClient) ListPodMetrics(namespace string) []PodMetrics {
	metricList := make([]PodMetrics, 0)
	pods, err := client.kubeClient.CoreV1().Pods(
		namespace,
	).List(
		context.TODO(),
		k8sapi.ListOptions{},
	)
	if err != nil {
		panic(err)
	}

	restConfig := getRestConfig()
	for _, pod := range pods.Items {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			containerStarted := *containerStatus.Started
			if !containerStarted {
				continue
			}

			err = execCmd(
				client.kubeClient,
				restConfig,
				pod.Name,
				"echo test",
				os.Stdin,
				os.Stdout,
				os.Stderr,
			)
			if err != nil {
				panic(err)
			}
		}
	}
	return metricList
}

// GetPodMetrics returns all metrics for a pod including memory and cpu usage
func (client *DAGMetricsClient) GetPodMetrics(namespace, name string) PodMetrics {
	// metrics, err := client.kubeClient.MetricsV1beta1().PodMetricses(
	// 	namespace,
	// ).Get(
	// 	context.TODO(),
	// 	name,
	// 	k8sapi.GetOptions{},
	// )
	// if err != nil {
	// 	panic(err)
	// }
	metrics := PodMetrics{}
	return metrics
}

// // GetPodMemory returns the current memory usage of the given pod
// func (client *DAGMetricsClient) GetPodMemory() {
// 	return
// }

// // GetPodCPU returns the current CPU usage of the given pod
// func (client *DAGMetricsClient) GetPodCPU(namespace, name string) int {
// 	return 0
// }
