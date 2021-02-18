package rest

import (
	"encoding/json"
	"fmt"
	"goflow/internal/config"
	dagconfig "goflow/internal/dag/config"
	"goflow/internal/dag/dagtype"
	"goflow/internal/dag/orchestrator"
	dagrun "goflow/internal/dag/run"
	"goflow/internal/testutils"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"net/http"

	"github.com/google/go-cmp/cmp"
	"k8s.io/client-go/kubernetes/fake"
)

var host string
var port int
var orch *orchestrator.Orchestrator
var testDag dagtype.DAG
var testTime time.Time
var testRun *dagrun.DAGRun
var goflowConfig *config.GoFlowConfig

func getTestOrchestrator(configuration *config.GoFlowConfig) *orchestrator.Orchestrator {
	kubeClient := fake.NewSimpleClientset()
	configuration.DAGPath = testutils.GetDagsFolder()
	configuration.DatabaseDNS = testutils.GetSQLiteLocation()
	return orchestrator.NewOrchestratorFromClientAndConfig(kubeClient, configuration)
}

func TestMain(m *testing.M) {
	configPath := testutils.GetConfigPath()
	goflowConfig = config.CreateConfig(configPath)
	host = "localhost"
	port = 8080
	testutils.RemoveSQLiteDB()
	orch = getTestOrchestrator(goflowConfig)
	testDag = dagtype.DAG{
		Config: &dagconfig.DAGConfig{
			Name: "test",
		},
	}
	testTime = time.Now()
	orch.AddDAG(&testDag)
	orch.GetDag(testDag.Config.Name).AddDagRun(testTime, false, nil)
	testRun = orch.GetDag(testDag.Config.Name).DAGRuns[0]
	go Serve(host, port, orch)
	m.Run()
}

func getUrl(suffix string) string {
	return fmt.Sprintf("http://%s:%d/%s", host, port, suffix)
}

func put(suffix string, content string) []byte {
	url := getUrl(suffix)
	reader := strings.NewReader(content)
	resp, err := http.Post(url, "json", reader)
	if err != nil {
		panic(err)
	}
	readBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return readBytes
}

func get(suffix string) []byte {
	url := getUrl(suffix)
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

// func getDagsPath() string {
// 	goflowConfig := config.GoFlowConfig{}
// 	fmt.Println("Path", configPath)
// 	configBytes, err := ioutil.ReadFile(configPath)
// 	if err != nil {
// 		panic(err)
// 	}
// 	json.Unmarshal(configBytes, goflowConfig)
// 	fmt.Println("Config", goflowConfig)
// 	return goflowConfig.DAGPath
// }

func TestPutDag(t *testing.T) {
	config := dagconfig.DAGConfig{
		Name:          "test-dag-4",
		Command:       []string{"echo", "1"},
		StartDateTime: "2019-01-01",
		EndDateTime:   "2020-01-01",
		Schedule:      "* * * * *",
	}
	configBytes, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}
	put("dag", string(configBytes))
	addedDagPath := path.Join(goflowConfig.DAGPath, fmt.Sprintf("%s.json", config.Name))
	fileBytes, err := ioutil.ReadFile(addedDagPath)
	if err != nil {
		panic(err)
	}

	dagConfigSeen := dagconfig.DAGConfig{}
	err = json.Unmarshal(fileBytes, &dagConfigSeen)
	if err != nil {
		panic(err)
	}
	defer os.Remove(addedDagPath)
	if ! cmp.Equal(config, dagConfigSeen) {
		t.Errorf("Expected config: \n%s\nbut generated config:\n%s", fmt.Sprint(&config), fmt.Sprint(&dagConfigSeen))
	}
}