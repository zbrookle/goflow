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

func getFakePod(name string) *core.Pod {
	container := getFakeContainerStatus()
	return &core.Pod{
		ObjectMeta: k8sapi.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Status: core.PodStatus{
			ContainerStatuses: []core.ContainerStatus{container},
		},
	}
}

func TestMain(m *testing.M) {
	fakePod = getFakePod("test1")
	fakePod2 := getFakePod("test2")
	fakeMetricsClient := fake.NewSimpleClientset(fakePod, fakePod2)
	metricsClient = NewDAGMetricsClient(fakeMetricsClient, true)
	m.Run()
}

// This is really just to ensure the code doesn't fail in test mode
func TestGetTestPodMetrics(t *testing.T) {
	_, err := metricsClient.GetPodMetrics(*fakePod)
	if err != nil {
		panic(err)
	}
}

func TestGetAllMetrics(t *testing.T) {
	metrics := metricsClient.ListPodMetrics(fakePod.Namespace)
	if len(metrics) != 2 {
		t.Error("Only expected two pods worth of metrics")
	}
}
