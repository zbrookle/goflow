package rest

import (
	"encoding/json"
	"fmt"
	"goflow/internal/dag/config"
	"goflow/internal/dag/dagtype"
	"goflow/internal/dag/orchestrator"
	dagrun "goflow/internal/dag/run"
	"goflow/internal/database"
	"goflow/internal/testutils"
	"io/ioutil"
	"testing"
	"time"

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
var testTime time.Time
var testRun *dagrun.DAGRun

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
	testTime = time.Now()
	orch.AddDAG(&testDag)
	orch.GetDag(testDag.Config.Name).AddDagRun(testTime, false, nil)
	testRun = orch.GetDag(testDag.Config.Name).DAGRuns[0]
	go serveSingle(host, port, orch, registerGetHandles)
	m.Run()
}

func get(suffix string) []byte {
	url := fmt.Sprintf("http://%s:%d/%s", host, port, suffix)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	readBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return readBytes
}

func TestGetDags(t *testing.T) {
	readBytes := get("dags")
	dagList := make([]dagtype.DAG, 0)
	json.Unmarshal(readBytes, &dagList)
	expectedDag := dagList[0]
	if expectedDag.Config.Name != testDag.Config.Name {
		t.Errorf("Expected dag not found!")
	}
}

func TestGetDag(t *testing.T) {
	bytes := get(fmt.Sprintf("dag/%s", testDag.Config.Name))
	dag := dagtype.DAG{}
	json.Unmarshal(bytes, &dag)
	if dag.Config.Name != testDag.Config.Name {
		t.Errorf("Expected dag with name %s", testDag.Config.Name)
	}
}

func TestGetMissingDag(t *testing.T) {
	bytes := get(fmt.Sprintf("dag/%s", "fake_dag"))
	if string(bytes) != missingDagMsg {
		t.Error("Message should indicate that DAG does not exist")
	}
}

func TestGetDagRuns(t *testing.T) {
	bytes := get(fmt.Sprintf("dag/%s/runs", testDag.Config.Name))
	dagRuns := make([]dagrun.DAGRun, 0)
	json.Unmarshal(bytes, &dagRuns)
	dagRun := dagRuns[0]
	if dagRun.Name != testRun.Name {
		t.Error("Expected dag run does not match")
	}
}
