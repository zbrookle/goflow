package metrics

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metrictype "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	k8smetrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

// DAGMetricsClient handles all interactions with DAG metrics
type DAGMetricsClient struct {
	k8sMetricsClient k8smetrics.Interface
}

// NewDAGMetricsClient returns a new DAGMetricsClient from a metrics clientset
func NewDAGMetricsClient(clientSet k8smetrics.Interface) *DAGMetricsClient {
	return &DAGMetricsClient{clientSet}
}

// GetPodMetrics returns all metrics for a pod including memory and cpu usage
func (client *DAGMetricsClient) GetPodMetrics(namespace, name string) *metrictype.PodMetrics {
	metrics, err := client.k8sMetricsClient.MetricsV1beta1().PodMetricses(
		namespace,
	).Get(
		context.TODO(),
		name,
		v1.GetOptions{},
	)
	if err != nil {
		panic(err)
	}
	return metrics
}

// GetPodMemory returns the current memory usage of the given pod
func (client *DAGMetricsClient) GetPodMemory() {
	return
}

// GetPodCPU returns the current CPU usage of the given pod
func (client *DAGMetricsClient) GetPodCPU(namespace, name string) int {
	return 0
}
