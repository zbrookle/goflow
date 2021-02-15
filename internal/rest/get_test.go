package rest

import (
	"encoding/json"
	"fmt"
	"goflow/internal/dag/config"
	"goflow/internal/dag/dagtype"
	"goflow/internal/dag/orchestrator"
	"goflow/internal/database"
	"goflow/internal/testutils"
	"io/ioutil"
	"testing"

	"net/http"

	"k8s.io/client-go/kubernetes/fake"
)

var kubeClient *fake.Clientset
var configPath string
var dagPath string
var sqlClient *database.SQLClient
var host string
var port int
var orch *orchestrator.Orchestrator
var testDag dagtype.DAG

func createFakeKubeClient() *fake.Clientset {
	return fake.NewSimpleClientset()
}

func TestMain(m *testing.M) {
	kubeClient = createFakeKubeClient()
	configPath = testutils.GetConfigPath()
	dagPath = testutils.GetDagsFolder()
	host = "localhost"
	port = 8080
	testutils.RemoveSQLiteDB()
	sqlClient = database.NewSQLiteClient(testutils.GetSQLiteLocation())
	orch = orchestrator.NewOrchestrator(configPath)
	testDag = dagtype.DAG{
		Config: &config.DAGConfig{
			Name: "test",
		},
	}
	orch.AddDAG(&testDag)
	go serveSingle(host, port, orch, registerGetHandles)
	m.Run()
}

func getURL(suffix string) string {
	return fmt.Sprintf("http://%s:%d/%s", host, port, suffix)
}

func TestGetDags(t *testing.T) {
	resp, err := http.Get(getURL("dags"))
	if err != nil {
		panic(err)
	}
	readBytes, err := ioutil.ReadAll(resp.Body)
	dagList := make([]dagtype.DAG, 0)
	json.Unmarshal(readBytes, &dagList)
	expectedDag := dagList[0]
	if expectedDag.Config.Name != testDag.Config.Name {
		t.Errorf("Expected dag not found!")
	}
}

// func TestGetDag(t *testing.T) {

// }

// func TestGetDagRuns(t *testing.T) {

// }
