package orchestrator

import (
	"testing"

	"goflow/dags"

	"k8s.io/client-go/kubernetes/fake"
)

var kubeClient *fake.Clientset

func createFakeKubeClient() *fake.Clientset {
	return fake.NewSimpleClientset()
}

func TestMain(m *testing.M) {
	kubeClient = createFakeKubeClient()
	m.Run()
}

func TestRegisterDAG(t *testing.T) {
	dag := dags.NewDAG("test", "default", "* * * * *", "busyboxy", "Never", kubeClient)
	orch := newOrchestratorFromClient(kubeClient)
	const expectedLength = 1
	orch.registerDag(dag)
	if orch.dagMap[dag.Name] != dag {
		t.Error("CronJob not added at correct key")
	}
	if len(orch.dagMap) != expectedLength {
		t.Errorf("CronMap should have length %d", expectedLength)
	}
}

func TestCollectDags(t *testing.T) {

}
