package servertest

import (
	"context"
	"encoding/json"
	"fmt"
	"goflow/config"
	"goflow/dagconfig"

	"goflow/k8sclient"
	"goflow/orchestrator"
	"goflow/podutils"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	newConfigPath := filepath.Join(podutils.GetTestFolder(), "tmp_config.json")
	fixedConfig.SaveConfig(newConfigPath)
	return newConfigPath
}

func getPods(kubeClient kubernetes.Interface) *[]*core.Pod {
	podSlice := make([]*core.Pod, 0)
	namespaces, err := kubeClient.CoreV1().Namespaces().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, namespace := range namespaces.Items {
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
			podSlice = append(podSlice, &pod)
		}
	}
	return &podSlice
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

func TestMain(m *testing.M) {
	expectedDagCount = 2
	fakeDagsPath := createFakeDags(podutils.GetTestFolder())
	defer os.RemoveAll(fakeDagsPath)
	configPath = adjustConfigDagPath(podutils.GetConfigPath(), fakeDagsPath)
	defer os.Remove(configPath)
	m.Run()
}

func TestStartServer(t *testing.T) {
	kubeClient := k8sclient.CreateKubeClient()
	defer podutils.CleanUpPods(kubeClient)
	orch := *orchestrator.NewOrchestrator(configPath)
	loopBreaker := make(chan struct{}, 1)
	orch.Start(1, loopBreaker)

	time.Sleep(4 * time.Second)
	loopBreaker <- struct{}{}

	if len(orch.DAGs()) != expectedDagCount {
		t.Errorf("Expected %d DAGs but only found %d", expectedDagCount, len(orch.DAGs()))
	}

	if len(orch.DagRuns()) == 0 {
		t.Error("Expected runs to present but none were found")
	}
}
