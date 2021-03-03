package testutils

// import (
// 	"context"
// 	"goflow/internal/dag/metrics"
// 	"time"

// 	v1 "k8s.io/api/core/v1"
// 	"k8s.io/apimachinery/pkg/api/resource"
// 	k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/runtime/schema"
// 	"k8s.io/client-go/kubernetes"

// 	metrictype "k8s.io/metrics/pkg/apis/metrics/v1beta1"
// 	fakemetrics "k8s.io/metrics/pkg/client/clientset/versioned/fake"
// )

// func registerPodsIntoMetrics(
// 	kubeClient kubernetes.Interface,
// 	metricsClient *fakemetrics.Clientset,
// ) {
// 	seenJobs := make(map[string]metrictype.PodMetrics)
// 	for true {
// 		namespaces, err := kubeClient.CoreV1().Namespaces().List(
// 			context.TODO(),
// 			k8sapi.ListOptions{},
// 		)
// 		if err != nil {
// 			panic(err)
// 		}
// 		for _, namespace := range namespaces.Items {
// 			podList, err := kubeClient.CoreV1().Pods(
// 				namespace.Name,
// 			).List(
// 				context.TODO(),
// 				k8sapi.ListOptions{},
// 			)
// 			if err != nil {
// 				panic(err)
// 			}
// 			for _, pod := range podList.Items {
// 				if _, ok := seenJobs[pod.Name]; ok {
// 					continue
// 				}
// 				containers := make([]metrictype.ContainerMetrics, 0)
// 				for _, container := range pod.Spec.Containers {
// 					containers = append(containers, metrictype.ContainerMetrics{
// 						Name: container.Name,
// 						Usage: map[v1.ResourceName]resource.Quantity{
// 							"": {
// 								Format: "",
// 							},
// 						},
// 					})
// 				}
// 				podMetrics := metrictype.PodMetrics{
// 					ObjectMeta: pod.ObjectMeta,
// 					Timestamp: k8sapi.Time{
// 						Time: time.Now(),
// 					},
// 					Window: k8sapi.Duration{
// 						Duration: 0,
// 					},
// 					Containers: containers,
// 				}
// 				seenJobs[pod.Name] = podMetrics
// 				metricsClient.Tracker().Create(schema.GroupVersionResource{
// 					Group:    "metrics.k8s.io",
// 					Version:  "v1beta1",
// 					Resource: "pods",
// 				}, &podMetrics, pod.Namespace)
// 			}
// 		}
// 	}
// }

// // NewTestMetricsClient returns a new metrics client for testing only
// func NewTestMetricsClient(interfaces ...kubernetes.Interface) *metrics.DAGMetricsClient {
// 	return metrics.NewDAGMetricsClient(interfaces[0])
// }
