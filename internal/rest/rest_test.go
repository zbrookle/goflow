package rest

import (
	"encoding/json"
	"fmt"
	"goflow/internal/config"
	dagconfig "goflow/internal/dag/config"
	"goflow/internal/dag/dagtype"
	"goflow/internal/dag/orchestrator"
	dagrun "goflow/internal/dag/run"
	dagtable "goflow/internal/dag/sql/dag"
	dagruntable "goflow/internal/dag/sql/dagrun"
	"goflow/internal/database"
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

func copyDAG(dag dagtype.DAG) dagtype.DAG {
	return dag
}

func TestMain(m *testing.M) {
	configPath := testutils.GetConfigPath()
	goflowConfig = config.CreateConfig(configPath)
	host = "localhost"
	port = 8080
	testutils.RemoveSQLiteDB()
	orch = getTestOrchestrator(goflowConfig)
	SQLCLIENT := database.NewSQLiteClient(testutils.GetSQLiteLocation())
	dagTableClient := dagtable.NewTableClient(SQLCLIENT)
	dagRunTableClient := dagruntable.NewTableClient(SQLCLIENT)
	dagTableClient.CreateTable()
	dagRunTableClient.CreateTable()
	testDag = dagtype.CreateDAG(&dagconfig.DAGConfig{
		Name:          "test",
		StartDateTime: "2019-01-01",
		MaxActiveRuns: 1,
	}, "", fake.NewSimpleClientset(), dagtype.ScheduleCache{}, dagTableClient, "", dagRunTableClient)
	testTime = time.Now()
	orch.AddDAG(&testDag)
	testDAG2 := copyDAG(testDag)
	testDAG2.Config.Name = "test2"
	orch.AddDAG(&testDAG2)
	orch.GetDag(testDag.Config.Name).AddDagRun(testTime, false, nil)
	testRun = orch.GetDag(testDag.Config.Name).DAGRuns[0]
	go Serve(host, port, orch)
	m.Run()
}

func getURL(suffix string) string {
	return fmt.Sprintf("http://%s:%d/%s", host, port, suffix)
}

func post(suffix string, content string) *http.Response {
	url := getURL(suffix)
	reader := strings.NewReader(content)
	resp, err := http.Post(url, "json", reader)
	if err != nil {
		panic(err)
	}
	return resp
}

func get(suffix string) *http.Response {
	url := getURL(suffix)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	return resp
}

func put(suffix string) *http.Response {
	url := getURL(suffix)
	client := &http.Client{}
	request, err := http.NewRequest("PUT", url, strings.NewReader(""))
	if err != nil {
		panic(err)
	}
	resp, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	return resp
}

func readRespBytes(resp *http.Response) []byte {
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return bodyBytes
}

func errorCodeResponse(t *testing.T, expectedCode int, received int) {
	if received != expectedCode {
		t.Errorf(
			"Should have code: %d, but received: %d",
			expectedCode,
			received,
		)
	}
}

func TestGetDags(t *testing.T) {
	resp := get("dags")
	bodyBytes := readRespBytes(resp)
	dagList := make([]dagtype.DAG, 0)
	err := json.Unmarshal(bodyBytes, &dagList)
	if err != nil {
		panic(err)
	}
	expectedDag := dagList[0]
	if expectedDag.Config.Name != testDag.Config.Name {
		t.Errorf("Expected dag not found!")
	}
	errorCodeResponse(t, http.StatusOK, resp.StatusCode)
}

func TestGetDag(t *testing.T) {
	resp := get(fmt.Sprintf("dag/%s", testDag.Config.Name))
	bodyBytes := readRespBytes(resp)
	dag := dagtype.DAG{}
	json.Unmarshal(bodyBytes, &dag)
	if dag.Config.Name != testDag.Config.Name {
		t.Errorf("Expected dag with name %s", testDag.Config.Name)
	}
	errorCodeResponse(t, http.StatusOK, resp.StatusCode)
}

func TestGetMissingDag(t *testing.T) {
	resp := get(fmt.Sprintf("dag/%s", "fake_dag"))
	bodyBytes := readRespBytes(resp)
	if string(bodyBytes) != missingDagMsg {
		t.Error("Message should indicate that DAG does not exist")
	}
	errorCodeResponse(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetDagRuns(t *testing.T) {
	resp := get(fmt.Sprintf("dag/%s/runs", testDag.Config.Name))
	bodyBytes := readRespBytes(resp)
	dagRuns := make([]dagrun.DAGRun, 0)
	json.Unmarshal(bodyBytes, &dagRuns)
	dagRun := dagRuns[0]
	if dagRun.Name != testRun.Name {
		t.Error("Expected dag run does not match")
	}
	errorCodeResponse(t, http.StatusOK, resp.StatusCode)
}

func TestPostDag(t *testing.T) {
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
	resp := post("dag", string(configBytes))
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
	if !cmp.Equal(config, dagConfigSeen) {
		t.Errorf(
			"Expected config: \n%s\nbut generated config:\n%s",
			fmt.Sprint(&config),
			fmt.Sprint(&dagConfigSeen),
		)
	}
	errorCodeResponse(t, http.StatusOK, resp.StatusCode)
}

func TestPostInvalidDag(t *testing.T) {
	config := dagconfig.DAGConfig{
		Name:          "test-dag-4.json",
		Command:       []string{"echo", "1"},
		StartDateTime: "2019-01-01",
		EndDateTime:   "2020-01-01",
		Schedule:      "* * * * *",
	}
	configBytes, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}
	resp := post("dag", string(configBytes))

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	if !strings.Contains(string(bodyBytes), "DAG name must match") {
		t.Error("Error response should have been raised!")
	}
	errorCodeResponse(t, http.StatusBadRequest, resp.StatusCode)
}

func TestToggleDag(t *testing.T) {
	orch.AddDAG(&testDag)
	path := fmt.Sprintf("dag/%s/toggle", testDag.Config.Name)
	put(path)
	if !testDag.IsOn {
		t.Error("DAG should be on!")
	}
	put(path)
	if testDag.IsOn {
		t.Error("DAG should be off!")
	}
}
