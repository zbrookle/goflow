package metrics

import (
	"context"
	"fmt"

	k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// DAGMetricsClient handles all interactions with DAG metrics
type DAGMetricsClient struct {
	kubeClient kubernetes.Interface
}

type PodMetrics struct {
	Memory int
	CPU    float32
}

// NewDAGMetricsClient returns a new DAGMetricsClient from a metrics clientset
func NewDAGMetricsClient(clientSet kubernetes.Interface) *DAGMetricsClient {
	return &DAGMetricsClient{clientSet}
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
	for _, pod := range pods.Items {
		fmt.Println(pod.Status)
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
