package orchestrator

import (
	"goflow/testutils"
	"testing"

	"goflow/config"
	"goflow/dags"

	"k8s.io/client-go/kubernetes/fake"
)

var kubeClient *fake.Clientset
var configPath string
var dagPath string

func createFakeKubeClient() *fake.Clientset {
	return fake.NewSimpleClientset()
}

func TestMain(m *testing.M) {
	kubeClient = createFakeKubeClient()
	configPath = testutils.GetConfigPath()
	dagPath = testutils.GetDagsFolder()
	m.Run()
}

func testOrchestrator() *Orchestrator {
	configuration := config.CreateConfig(configPath)
	configuration.DAGPath = dagPath
	return newOrchestratorFromClientAndConfig(kubeClient, configuration)
}

func TestRegisterDAG(t *testing.T) {
	dag := dags.NewDAG("test", "default", "* * * * *", "busyboxy", "Never", kubeClient)
	orch := testOrchestrator()
	const expectedLength = 1
	orch.AddDAG(dag)
	if orch.dagMap[dag.Name] != dag {
		t.Error("DAG not added at correct key")
	}
	if len(orch.dagMap) != expectedLength {
		t.Errorf("DAG map should have length %d", expectedLength)
	}
}

func TestCollectDags(t *testing.T) {
	orch := testOrchestrator()
	orch.CollectDAGs()
	dagCount := len(orch.DAGs())
	if dagCount == 0 {
		t.Errorf("%d DAGs collected, expected more than 0", dagCount)
	}
	for dagName := range orch.dagMap {
		if dagName != orch.dagMap[dagName].Name {
			panic("Key doesn't match up with dag name!")
		}
	}
}
