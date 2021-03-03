package metrics

import (
	// "context"
	// "fmt"
	"testing"

	core "k8s.io/api/core/v1"
	k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var metricsClient *DAGMetricsClient
var fakePod *core.Pod
var fakeContainer *core.Container

func getFakeContainerStatus() core.ContainerStatus {
	started := true
	return core.ContainerStatus{
		Name:    "task",
		Ready:   true,
		Started: &started,
	}
}

func getFakePod() *core.Pod {
	container := getFakeContainerStatus()
	return &core.Pod{
		ObjectMeta: k8sapi.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Status: core.PodStatus{
			ContainerStatuses: []core.ContainerStatus{container},
		},
	}
}

func TestMain(m *testing.M) {
	fakePod = getFakePod()
	fakeMetricsClient := fake.NewSimpleClientset(fakePod)
	metricsClient = NewDAGMetricsClient(fakeMetricsClient, true)
	m.Run()
}

func TestGetAllPodMetrics(t *testing.T) {
	metrics, err := metricsClient.GetPodMetrics(*fakePod)
	if err != nil {
		panic(err)
	}
	t.Error(metrics)
}
