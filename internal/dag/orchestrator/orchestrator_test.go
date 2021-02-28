package orchestrator

import (
	dagconfig "goflow/internal/dag/config"
	"goflow/internal/database"
	"goflow/internal/testutils"
	"testing"

	"goflow/internal/config"
	"goflow/internal/dag/dagtype"

	"k8s.io/client-go/kubernetes/fake"
)

var kubeClient *fake.Clientset
var configPath string
var dagPath string
var sqlClient *database.SQLClient

const newImageName = "differentImage"

func createFakeKubeClient() *fake.Clientset {
	return fake.NewSimpleClientset()
}

func TestMain(m *testing.M) {
	kubeClient = createFakeKubeClient()
	configPath = testutils.GetConfigPath()
	dagPath = testutils.GetDagsFolder()
	testutils.RemoveSQLiteDB()
	sqlClient = database.NewSQLiteClient(testutils.GetSQLiteLocation())
	m.Run()
}

func testOrchestrator() *Orchestrator {
	configuration := config.CreateConfig(configPath)
	configuration.DAGPath = dagPath
	configuration.DatabaseDNS = testutils.GetSQLiteLocation()
	return NewOrchestratorFromClientsAndConfig(
		kubeClient,
		configuration,
		testutils.NewTestMetricsClient(),
	)
}

func getTestDAG(orch *Orchestrator) dagtype.DAG {
	config := &dagconfig.DAGConfig{
		Name:          "test",
		Namespace:     "default",
		Schedule:      "* * * * *",
		DockerImage:   "busybox",
		RetryPolicy:   "Never",
		Command:       []string{"echo", "yes"},
		TimeLimit:     nil,
		MaxActiveRuns: 1,
		StartDateTime: "2019-01-01",
		EndDateTime:   "",
	}
	return dagtype.CreateDAG(
		config,
		config.String(),
		orch.kubeClient,
		orch.metricsClient,
		make(dagtype.ScheduleCache),
		orch.dagTableClient,
		"path",
		orch.dagrunTableClient,
		true,
	)
}

func TestRegisterDAG(t *testing.T) {
	defer database.PurgeDB(sqlClient)
	orch := testOrchestrator()
	orch.dagTableClient.CreateTable()
	dag := getTestDAG(orch)
	const expectedLength = 1
	orch.AddDAG(&dag)
	if orch.dagMap[dag.Config.Name] != &dag {
		t.Error("DAG not added at correct key")
	}
	if len(orch.dagMap) != expectedLength {
		t.Errorf("DAG map should have length %d", expectedLength)
	}
}

func getDagWithDifferentDockerImage(orch *Orchestrator) dagtype.DAG {
	updatedDAG := getTestDAG(orch)
	newConfig := updatedDAG.Config.Copy()
	newConfig.DockerImage = newImageName
	updatedDAG.Config = &newConfig
	updatedDAG.Code = newConfig.String()
	return updatedDAG
}

func TestDAGUpdate(t *testing.T) {
	defer database.PurgeDB(sqlClient)
	orch := testOrchestrator()
	orch.dagTableClient.CreateTable()
	dag := getTestDAG(orch)
	orch.AddDAG(&dag)
	updatedDAG := getDagWithDifferentDockerImage(orch)
	orch.UpdateDag(&updatedDAG)
	if orch.dagMap[dag.Config.Name].Config.DockerImage != newImageName {
		t.Errorf("Expected image name to be updated to \"%s\"", newImageName)
	}
}

func TestCollectDagUpdatedTime(t *testing.T) {
	defer database.PurgeDB(sqlClient)
	orch := testOrchestrator()
	orch.dagTableClient.CreateTable()
	dag := getTestDAG(orch)
	orch.collectDAG(&dag)
	addedTime := dag.LastUpdated
	updatedDAG := getDagWithDifferentDockerImage(orch)
	orch.collectDAG(&updatedDAG)
	foundLastUpdateTime := orch.dagMap[dag.Config.Name].LastUpdated
	if !foundLastUpdateTime.After(addedTime) {
		t.Errorf(
			"Expected dag to have later update time than '%s', but found '%s'",
			addedTime,
			foundLastUpdateTime,
		)
	}
}

func TestUpdateDAGWhileRunning(t *testing.T) {
	defer database.PurgeDB(sqlClient)
	orch := testOrchestrator()
	orch.dagTableClient.CreateTable()
	orch.dagrunTableClient.CreateTable()
	dag := getTestDAG(orch)
	dag.IsOn = true
	orch.collectDAG(&dag)
	dag.AddNextDagRunIfReady(orch.channelHolder)
	updatedDAG := getDagWithDifferentDockerImage(orch)
	updatedDAG.IsOn = true
	orch.collectDAG(&updatedDAG)
	retrievedDAG := orch.dagMap[dag.Config.Name]
	if retrievedDAG.Config.DockerImage != newImageName {
		t.Errorf("Expected image name to be updated to \"%s\"", newImageName)
	}
}

func TestCollectDags(t *testing.T) {
	defer database.PurgeDB(sqlClient)
	orch := testOrchestrator()
	orch.dagTableClient.CreateTable()
	orch.CollectDAGs()
	dagCount := len(orch.DAGs())
	if dagCount == 0 {
		t.Errorf("%d DAGs collected, expected more than 0", dagCount)
	}
	for dagName := range orch.dagMap {
		if dagName != orch.dagMap[dagName].Config.Name {
			panic("Key doesn't match up with dag name!")
		}
	}
}
