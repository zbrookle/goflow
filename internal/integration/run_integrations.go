package main

import (
	"context"
	"encoding/json"
	"fmt"
	"goflow/internal/config"
	dagconfig "goflow/internal/dag/config"
	"goflow/internal/dag/dagtype"
	"goflow/internal/logs"

	"goflow/internal/dag/orchestrator"
	k8sclient "goflow/internal/k8s/client"

	podutils "goflow/internal/k8s/pod/utils"
	"goflow/internal/testutils"
	"io/ioutil"
	"os"
	"path/filepath"

	"regexp"
	"strconv"
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

func getLogMessage(num int) string {
	return fmt.Sprintf("Hello world %d", num)
}

func createFakeDagFile(dagFolder string, dagNum int) {
	fakeDagName := fmt.Sprintf("dag_file_%d", dagNum)
	filePath := filepath.Join(dagFolder, fakeDagName+".json")
	echoString := fmt.Sprintf("echo \"%s\"", getLogMessage(dagNum))
	fakeDagConfig := &dagconfig.DAGConfig{Name: fakeDagName,
		Namespace:     "default",
		Schedule:      "* * * * *",
		DockerImage:   "busybox",
		RetryPolicy:   core.RestartPolicyOnFailure,
		Command:       []string{"sh", "-c", echoString},
		Parallelism:   0,
		TimeLimit:     nil,
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

func getDagID(config dagconfig.DAGConfig) int {
	re := regexp.MustCompile("dag_file_(\\d)")
	matchGroups := re.FindStringSubmatch(config.Name)
	id, err := strconv.Atoi(matchGroups[1])
	if err != nil {
		panic(err)
	}
	return id
}

func podNamesInPodMap(names map[string]struct{}, podMap map[string]*core.Pod) bool {
	for name := range names {
		_, ok := podMap[name]
		if ok {
			return true
		}
	}
	return false
}

func waitUntilPodsGoneOrTimePassed(
	seconds time.Duration,
	client kubernetes.Interface,
	firstDAGRunPodNamesSet map[string]struct{},
) {
	logs.InfoLogger.Println(firstDAGRunPodNamesSet)
	endWait := make(chan struct{})

	secondsTime := seconds * time.Second
	go func() {
		time.Sleep(secondsTime)
		panic(fmt.Sprintf("%d seconds have passed, pods still not deleted\n", seconds))
	}()

	// Wait for pods deleted
	go func() {
		for {
			podMap := getPods(client)
			defaultNameSpaceMap := podMap["default"]
			select {
			case _, _ = <-endWait:
				break
			default:
				if !podNamesInPodMap(firstDAGRunPodNamesSet, defaultNameSpaceMap) {
					logs.InfoLogger.Println("Pods gone")
					close(endWait)
					return
				}
				time.Sleep(time.Second)
			}
		}

	}()

	_, _ = <-endWait
	logs.InfoLogger.Println("Wait done")
}

func getDateFromString(dateStr string) time.Time {
	time, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		panic(err)
	}
	return time
}

func getFirstRunDagNames(dags []*dagtype.DAG) map[string]struct{} {
	names := make(map[string]struct{})
	for _, dag := range dags {
		dagFirstRunDateString := getDateFromString(dag.Config.StartDateTime).String()
		dagName := podutils.CleanK8sName(dag.Config.Name + "-" + dagFirstRunDateString)
		names[dagName] = struct{}{}
	}
	return names
}

func startServer() {
	kubeClient := k8sclient.CreateKubeClient()
	defer podutils.CleanUpEnvironment(kubeClient)
	orch := *orchestrator.NewOrchestrator(configPath)
	go orch.Start(1 * time.Second)

	time.Sleep(3 * time.Second)

	dags := orch.DAGs()
	if len(dags) != expectedDagCount {
		panic(fmt.Sprintf("Expected %d DAGs but only found %d", expectedDagCount, len(dags)))
	}

	if len(orch.DagRuns()) == 0 {
		panic("Expected runs to be present but none were found")
	}

	firstRunDagNames := getFirstRunDagNames(dags)
	waitUntilPodsGoneOrTimePassed(30, kubeClient, firstRunDagNames)

	for _, run := range orch.DagRuns() {
		logs.InfoLogger.Println(run)
		_, ok := firstRunDagNames[run.Name]
		if ok {
			select {
			case logText := <-run.Logs():
				withoutNewlines := strings.TrimSpace(logText)
				expectedLogMessage := getLogMessage(getDagID(*run.Config))
				if withoutNewlines != expectedLogMessage {
					panic(
						fmt.Sprintf(
							"Expected log message %s, but got message %s",
							expectedLogMessage,
							withoutNewlines,
						),
					)
				}
			default:
				logs.InfoLogger.Println(run.Logs())
				panic(fmt.Sprintf("No logs available for pod %s!!!", run.Name))
			}
		}

	}

	orch.Stop()
}

func init() {
	logs.InfoLogger.Println("Starting goflow simulation program...")
	expectedDagCount = 6
}

func main() {
	fakeDagsPath := createFakeDags(testutils.GetTestFolder())
	defer os.RemoveAll(fakeDagsPath)
	configPath = adjustConfigDagPath(testutils.GetConfigPath(), fakeDagsPath)
	defer os.Remove(configPath)
	startServer()
}
