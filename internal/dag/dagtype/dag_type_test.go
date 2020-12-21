package dagtype

import (
	"context"
	goflowconfig "goflow/internal/config"
	"goflow/internal/dag/activeruns"
	dagconfig "goflow/internal/dag/config"
	dagrun "goflow/internal/dag/run"
	"goflow/internal/database"
	k8sclient "goflow/internal/k8s/client"
	"goflow/internal/k8s/pod/event/holder"
	podutils "goflow/internal/k8s/pod/utils"
	"goflow/internal/testutils"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	"encoding/json"

	dagtable "goflow/internal/dag/sql/dag"

	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var DAGPATH string
var KUBECLIENT kubernetes.Interface
var TABLECLIENT *dagtable.TableClient
var SQLCLIENT *database.SQLClient

func setUpNamespaces(client kubernetes.Interface) {
	namespaceClient := client.CoreV1().Namespaces()
	for _, name := range []string{"default"} {
		namespaceClient.Create(
			context.TODO(),
			&core.Namespace{ObjectMeta: v1.ObjectMeta{Name: name}},
			v1.CreateOptions{},
		)
	}
}

func TestMain(m *testing.M) {
	DAGPATH = filepath.Join(testutils.GetTestFolder(), "test_dags")
	KUBECLIENT = fake.NewSimpleClientset()

	testutils.RemoveSQLiteDB()
	SQLCLIENT = database.NewSQLiteClient(testutils.GetSQLiteLocation())
	TABLECLIENT = dagtable.NewTableClient(SQLCLIENT)
	podutils.CleanUpEnvironment(KUBECLIENT)
	setUpNamespaces(KUBECLIENT)
	m.Run()
}

type StringMap map[string]string

func map1InMap2(map1 StringMap, map2 StringMap) bool {
	for str := range map1 {
		if map1[str] != map2[str] {
			return false
		}
	}
	return true
}

func (stringMap StringMap) Equals(otherMap StringMap) bool {
	return map1InMap2(stringMap, otherMap) && map1InMap2(otherMap, stringMap)
}

func (stringMap StringMap) Bytes() []byte {
	bytes, err := json.Marshal(stringMap)
	if err != nil {
		panic(err)
	}
	return bytes
}

func setUpDatabase() {
	TABLECLIENT.CreateTable()
}

func TestDAGFromJSONBytes(t *testing.T) {
	defer database.PurgeDB(SQLCLIENT)
	setUpDatabase()
	config := dagconfig.DAGConfig{
		Name:          "test",
		Namespace:     "default",
		Schedule:      "* * * * *",
		DockerImage:   "busybox",
		RetryPolicy:   core.RestartPolicyNever,
		Command:       []string{"echo", "yes"},
		Parallelism:   1,
		TimeLimit:     nil,
		Retries:       int32(2),
		StartDateTime: "2019-01-01",
		EndDateTime:   "2020-01-01",
		Labels:        map[string]string{"test": "test-label"},
		Annotations:   map[string]string{"anno": "value"},
		MaxActiveRuns: 1,
	}
	formattedJSONString := string(config.Marshal())
	expectedDAG := DAG{
		Config:              &config,
		Code:                string(config.Marshal()),
		StartDateTime:       time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDateTime:         time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		DAGRuns:             make([]*dagrun.DAGRun, 0),
		kubeClient:          nil,
		ActiveRuns:          activeruns.New(),
		MostRecentExecution: time.Time{},
		timeLock:            &sync.Mutex{},
	}
	expectedJSONString := string(expectedDAG.Marshal())
	dag, err := createDAGFromJSONBytes(
		[]byte(formattedJSONString),
		fake.NewSimpleClientset(),
		goflowconfig.GoFlowConfig{},
		make(ScheduleCache),
		TABLECLIENT,
		"path",
	)
	if err != nil {
		panic(err)
	}
	marshaledJSON := string(dag.Marshal())
	if expectedJSONString != marshaledJSON {
		t.Error("DAG struct does not match up with expected values")
		t.Error("Found:", dag)
		t.Error("Expec:", expectedDAG)
	}
}

func TestReadFiles(t *testing.T) {
	expectedFiles := []string{"my_json_dag.json", "my_json_dag2.json", "my_python_dag.py"}
	sort.Strings(expectedFiles)
	foundFilePaths := getDirSliceRecur(DAGPATH)
	for i, filePath := range foundFilePaths {
		_, foundFilePaths[i] = filepath.Split(filePath)
	}
	sort.Strings(foundFilePaths)
	expectedFileCount := len(expectedFiles)
	foundFileCount := len(foundFilePaths)
	if len(expectedFiles) != len(foundFilePaths) {
		t.Errorf("Expected %d files, found %d files", expectedFileCount, foundFileCount)
		panic("File counts are different")
	}
	for i, foundPath := range foundFilePaths {
		expectedFile := expectedFiles[i]
		_, foundFile := filepath.Split(foundPath)
		if expectedFiles[i] != foundFile {
			t.Errorf("Expected file %s, found file %s", expectedFile, foundFile)
		}
	}
}

func getTestDAG(client kubernetes.Interface) *DAG {
	dag := CreateDAG(&dagconfig.DAGConfig{
		Name:          "test",
		Namespace:     "default",
		Schedule:      "* * * * *",
		DockerImage:   "busybox",
		RetryPolicy:   "Never",
		Command:       []string{"echo", "\"Hello world!!!!!!!\""},
		TimeLimit:     nil,
		MaxActiveRuns: 1,
		StartDateTime: "2019-01-01",
		EndDateTime:   "",
	}, "", client, make(ScheduleCache), TABLECLIENT, "path")
	return &dag
}

func getTestDAGFakeClient() *DAG {
	return getTestDAG(KUBECLIENT)
}

func getTestDAGRealClient() *DAG {
	return getTestDAG(k8sclient.CreateKubeClient())
}

func getTestDate() time.Time {
	return time.Date(2019, 1, 1, 0, 0, 0, 0, time.Now().Location())
}

func reportErrorCounts(t *testing.T, foundCount int, expectedCount int, testDag *DAG) {
	if foundCount != expectedCount {
		t.Errorf(
			"DAG Run not properly added, expected %d dag run, found %d",
			expectedCount,
			foundCount,
		)
		t.Error("Found dags:", testDag.DAGRuns)
	}
}

func TestAddDagRun(t *testing.T) {
	defer database.PurgeDB(SQLCLIENT)
	setUpDatabase()
	testDAG := getTestDAGFakeClient()
	currentTime := getTestDate()
	testDAG.AddDagRun(currentTime, testDAG.Config.WithLogs, holder.New())
	reportErrorCounts(t, len(testDAG.DAGRuns), 1, testDAG)
}

func TestAddDagRunIfReady(t *testing.T) {
	defer database.PurgeDB(SQLCLIENT)
	setUpDatabase()
	actionCases := []struct {
		actionFunc   func(dag *DAG)
		expectedRuns int
	}{
		{
			func(dag *DAG) {},
			1,
		},
		{
			func(dag *DAG) { dag.ActiveRuns.Dec() },
			2,
		},
	}

	for _, action := range actionCases {
		func() {
			defer podutils.CleanUpEnvironment(KUBECLIENT)
			testDAG := getTestDAGFakeClient()
			channelHolder := holder.New()
			testDAG.AddNextDagRunIfReady(channelHolder)
			action.actionFunc(testDAG)
			testDAG.AddNextDagRunIfReady(channelHolder)
			reportErrorCounts(t, len(testDAG.DAGRuns), action.expectedRuns, testDAG)
		}()
	}
}
