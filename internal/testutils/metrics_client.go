package testutils

import (
	"context"
	"fmt"
	"goflow/internal/dag/metrics"

	k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	fakemetrics "k8s.io/metrics/pkg/client/clientset/versioned/fake"
)

func registerPodsIntoMetrics(
	kubeClient *fake.Clientset,
	metricsClient *fakemetrics.Clientset,
) {
	namespaces, err := kubeClient.CoreV1().Namespaces().List(context.TODO(), k8sapi.ListOptions{})
	if err != nil {
		panic(err)
	}
	for true {
		for _, namespace := range namespaces.Items {
			podList, err := kubeClient.CoreV1().Pods(
				namespace.Name,
			).List(
				context.TODO(),
				k8sapi.ListOptions{},
			)
			if err != nil {
				panic(err)
			}
			fmt.Println(podList)
		}
	}
	// kubeClient.CoreV1().Pods()
	// metricsClient.Tracker().Add()
}

// NewTestMetricsClient returns a new metrics client for testing only
func NewTestMetricsClient(interfaces ...*fake.Clientset) *metrics.DAGMetricsClient {
	fakeMetricsClientSet := fakemetrics.NewSimpleClientset()
	for _, inter := range interfaces {
		go registerPodsIntoMetrics(inter, fakeMetricsClientSet)
	}
	return metrics.NewDAGMetricsClient(fakeMetricsClientSet)
}
