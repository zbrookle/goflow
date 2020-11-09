package dags

import (
	"context"
	"goflow/jsonpanic"
	"goflow/testpaths"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"encoding/json"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

var DAGPATH string
var KUBECLIENT kubernetes.Interface

func TestMain(m *testing.M) {
	DAGPATH = filepath.Join(testpaths.GetTestFolder(), "test_dags")
	KUBECLIENT = fake.NewSimpleClientset()
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

func TestDAGFromJSONBytes(t *testing.T) {
	name := "test"
	namespace := "default"
	schedule := "* * * * *"
	image := "busybox"
	retryPolicy := "Never"
	command := "echo yes"
	parallelism := int32(1)
	timeLimit := int64(300)
	retries := int32(2)
	labels, _ := json.Marshal(map[string]string{"test": "test-label"})
	annotations, _ := json.Marshal(map[string]string{"anno": "value"})
	formattedJSONString := fmt.Sprintf(
		"{\"Name\":\"%s\",\"Namespace\":\"%s\",\"Schedule\":\"%s\",\"DockerImage\":\"%s\","+
			"\"RetryPolicy\":\"%s\",\"Command\":\"%s\",\"Parallelism\":%d,\"TimeLimit\":%d,"+
			"\"Retries\":%d,\"Labels\":%s,\"Annotations\":%s",
		name,
		namespace,
		schedule,
		image,
		retryPolicy,
		command,
		parallelism,
		timeLimit,
		retries,
		labels,
		annotations,
	)
	expectedJSONString := formattedJSONString + ",\"DAGRuns\":[]}"
	dag, err := createDAGFromJSONBytes([]byte(formattedJSONString + "}"))
	if err != nil {
		panic(err)
	}
	marshaledJSON, err := json.Marshal(dag)
	if err != nil {
		panic(err)
	}
	marshaledJSONString := string(marshaledJSON)
	if expectedJSONString != marshaledJSONString {
		t.Error("DAG struct does not match up with expected values")
		t.Error("Found:", marshaledJSONString)
		t.Error("Expected:", expectedJSONString)
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

func getTestDag() *DAG {
	return NewDAG("test", "default", "* * * * *", "busybox", "Never", KUBECLIENT)
}

func getTestDate() time.Time {
	return time.Date(2019, 1, 1, 0, 0, 0, 0, time.Now().Location())
}

func TestAddDagRun(t *testing.T) {
	testDag := getTestDag()
	currentTime := getTestDate()
	testDag.AddDagRun(currentTime)
	foundDagCount := len(testDag.DAGRuns)
	expectedCount := 1
	if foundDagCount != expectedCount {
		t.Errorf(
			"DAG Run not properly added, expected %d dag run, found %d",
			expectedCount,
			foundDagCount,
		)
		t.Error("Found dags:", testDag.DAGRuns)
	}
}

func TestCreateJob(t *testing.T) {
	dagRun := createDagRun(getTestDate(), getTestDag())
	dagRun.CreateJob()
	foundJob, err := dagRun.DAG.kubeClient.BatchV1().Jobs(
		dagRun.DAG.Namespace,
	).Get(
		context.TODO(),
		dagRun.Job.Name,
		v1.GetOptions{},
	)
	if err != nil {
		panic(err)
	}
	foundJobValue := jsonpanic.JSONPanic(*foundJob)
	expectedValue := jsonpanic.JSONPanic(*dagRun.Job)
	if foundJobValue != expectedValue {
		t.Error("Expected:", expectedValue)
		t.Error("Found:", foundJobValue)
	}
}
