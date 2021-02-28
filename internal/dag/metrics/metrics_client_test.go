package metrics

import (
	// "context"
	// "fmt"
	"encoding/json"
	"testing"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	metricsapi "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"k8s.io/metrics/pkg/client/clientset/versioned/fake"
)

var metricsClient *DAGMetricsClient
var fakePodMetrics *metricsapi.PodMetrics
var fakeContainerMetrics *metricsapi.ContainerMetrics

func getFakeContainer() metricsapi.ContainerMetrics {
	return metricsapi.ContainerMetrics{
		Name: "testContainer",
		Usage: map[core.ResourceName]resource.Quantity{
			"": {
				Format: "",
			},
		},
	}
}

func getFakePod() *metricsapi.PodMetrics {
	container := getFakeContainer()
	return &metricsapi.PodMetrics{
		ObjectMeta: k8sapi.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		// Window: k8sapi.Duration{
		// 	Duration: 0,
		// },
		Containers: []metricsapi.ContainerMetrics{container},
	}
}

func TestMain(m *testing.M) {
	fakePodMetrics = getFakePod()
	fakeMetricsClient := fake.NewSimpleClientset(fakePodMetrics)
	fakeMetricsClient.Tracker().Create(
		schema.GroupVersionResource{
			Group:    "metrics.k8s.io",
			Version:  "v1beta1",
			Resource: "pods",
		},
		fakePodMetrics,
		"default",
	)
	metricsClient = NewDAGMetricsClient(fakeMetricsClient)
	m.Run()
}

func TestGetAllPodMetrics(t *testing.T) {
	metrics := metricsClient.GetPodMetrics(fakePodMetrics.Namespace, fakePodMetrics.Name)
	_, err := json.Marshal(metrics)
	if err != nil {
		panic(err)
	}
	// t.Error(string(metricsJSON))
}
