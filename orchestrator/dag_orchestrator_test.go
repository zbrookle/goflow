package orchestrator

import (
	"goflow/testpaths"
	"testing"

	"goflow/config"
	"goflow/dags"

	"k8s.io/client-go/kubernetes/fake"
)

var kubeClient *fake.Clientset
var configPath string

func createFakeKubeClient() *fake.Clientset {
	return fake.NewSimpleClientset()
}

func TestMain(m *testing.M) {
	kubeClient = createFakeKubeClient()
	configPath = testpaths.GetConfigPath()
	m.Run()
}

func TestRegisterDAG(t *testing.T) {
	dag := dags.NewDAG("test", "default", "* * * * *", "busyboxy", "Never", kubeClient)
	orch := newOrchestratorFromClientAndConfig(kubeClient, config.CreateConfig(configPath))
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
