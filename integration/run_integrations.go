package main

import (
	"context"
	"encoding/json"
	"fmt"
	"goflow/config"
	dagconfig "goflow/dag/config"
	"goflow/logs"

	"goflow/dag/orchestrator"
	k8sclient "goflow/k8s/client"
	podutils "goflow/k8s/pod/utils"
	"goflow/testutils"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var configPath string
var expectedDagCount int

func adjustConfigDagPath(configPath string, dagPath string) string {
	fixedConfig := &config.GoFlowConfig{}
	configBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(configBytes, fixedConfig)
	fixedConfig.DAGPath = dagPath
	newConfigPath := filepath.Join(testutils.GetTestFolder(), "tmp_config.json")
	fixedConfig.SaveConfig(newConfigPath)
	return newConfigPath
}

func getPods(kubeClient kubernetes.Interface) map[string]map[string]*core.Pod {
	podMap := make(map[string]map[string]*core.Pod)
	namespaces, err := kubeClient.CoreV1().Namespaces().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, namespace := range namespaces.Items {
		namespaceName := namespace.Name
		namespaceDict, ok := podMap[namespaceName]
		if !ok {
			podMap[namespaceName] = make(map[string]*core.Pod)
			namespaceDict = podMap[namespaceName]
		}
		podList, err := kubeClient.CoreV1().Pods(
			namespace.Name,
		).List(
			context.TODO(),
			v1.ListOptions{},
		)
		if err != nil {
			panic(err)
		}
		for _, pod := range podList.Items {
			namespaceDict[pod.Name] = &pod
		}
	}
	return podMap
}

func createDirIfNotExist(directory string) string {
	_, err := os.Stat(directory)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(directory, 0755)
		if errDir != nil {
			panic(err)
		}
	}
	return directory
}

func createFakeDagFile(dagFolder string, dagNum int) {
	fakeDagName := fmt.Sprintf("dag_file_%d", dagNum)
	filePath := filepath.Join(dagFolder, fakeDagName+".json")
	fakeDagConfig := &dagconfig.DAGConfig{Name: fakeDagName,
		Namespace:     "default",
		Schedule:      "* * * * *",
		Command:       strings.Split(fmt.Sprintf("echo %d", dagNum), " "),
		Parallelism:   0,
		TimeLimit:     0,
		Retries:       2,
		MaxActiveRuns: 1,
		StartDateTime: "2019-01-01",
		EndDateTime:   "2020-01-01",
		WithLogs:      true,
	}
	jsonContent := fakeDagConfig.Marshal()
	ioutil.WriteFile(filePath, jsonContent, 0755)
}

// createFakeDags creates fake dag files, and returns their location
func createFakeDags(testFolder string) string {
	dagDir := createDirIfNotExist(filepath.Join(testFolder, "tmp_dags"))
	for i := 0; i < expectedDagCount; i++ {
		createFakeDagFile(dagDir, i)
	}
	return dagDir
}

func startServer() {
	kubeClient := k8sclient.CreateKubeClient()
	defer podutils.CleanUpPods(kubeClient)
	orch := *orchestrator.NewOrchestrator(configPath)
	loopBreaker := make(chan struct{}, 1)
	go orch.Start(1, loopBreaker)

	time.Sleep(4 * time.Second)

	if len(orch.DAGs()) != expectedDagCount {
		panic(fmt.Sprintf("Expected %d DAGs but only found %d", expectedDagCount, len(orch.DAGs())))
	}

	if len(orch.DagRuns()) == 0 {
		panic("Expected runs to present but none were found")
	}

	// podsOnServer := getPods(kubeClient)
	// for _, run := range orch.DagRuns() {
	// 	mostRecent := run.MostRecentPod()

	// 	namespaceDict, ok := podsOnServer[mostRecent.Namespace]
	// 	if !ok {
	// 		panic(fmt.Sprintf("Namespace %s not found in map", mostRecent.Namespace))
	// 	}
	// 	_, ok = namespaceDict[mostRecent.Name]
	// 	if !ok {
	// 		panic(
	// 			fmt.Sprintf(
	// 				"Pod %s not found in namespace %s",
	// 				mostRecent.Name,
	// 				mostRecent.Namespace,
	// 			),
	// 		)
	// 	}

	// 	// select {
	// 	// case logString, ok := <-*run.Logs():
	// 	// 	if ok {
	// 	// 		logs.InfoLogger.Println(logString)
	// 	// 	} else {
	// 	// 		panic("Channel CLosed!")
	// 	// 	}
	// 	// default:
	// 	// 	panic("No logs found in channel!")
	// 	// }
	// }

	close(loopBreaker)
}

func init() {
	logs.InfoLogger.Println("Starting goflow simulation program...")
	expectedDagCount = 2
}

func main() {
	fakeDagsPath := createFakeDags(testutils.GetTestFolder())
	defer os.RemoveAll(fakeDagsPath)
	configPath = adjustConfigDagPath(testutils.GetConfigPath(), fakeDagsPath)
	defer os.Remove(configPath)
	startServer()
}

// func TestRunMultiplePodsAtOnce(t *testing.T) {
// 	// Test with logs and without logs
// 	realClient := k8sclient.CreateKubeClient()
// 	tables := []struct {
// 		name     string
// 		withLogs bool
// 	}{
// 		{"Without Logs", false},
// 		{"With Logs", true},
// 	}
// 	for _, table := range tables {
// 		t.Logf("Test case: %s", table.name)
// 		func() {
// 			defer podutils.CleanUpPods(realClient)
// 			dagRun := newDAGRun(getTestDate(), getTestDAGRealClient(), table.withLogs)
// 			dagRun.Start()

// 			// Test for dag completion in state of dag
// 			if (dagRun.watcher.Phase != core.PodSucceeded) &&
// 				(dagRun.watcher.Phase != core.PodFailed) {
// 				t.Errorf(
// 					"A finished dagRun should be in phase %s or state %s, but found in state %s",
// 					core.PodSucceeded,
// 					core.PodFailed,
// 					dagRun.watcher.Phase,
// 				)
// 			}

// 			// Test for log output if logs enabled
// 			if table.withLogs {
// 				logMsg := <-*dagRun.Logs()
// 				expectedLogMessage := dagRun.DAG.Config.Command[1]
// 				logMsg = strings.ReplaceAll(logMsg, "\n", "")
// 				if logMsg != expectedLogMessage {
// 					t.Errorf(
// 						"Expected log message %s, found log message %s",
// 						expectedLogMessage,
// 						logMsg,
// 					)
// 				}
// 			}
// 		}()

// 	}
// }
