package servertest

import (
	"encoding/json"
	"fmt"
	"goflow/config"
	"goflow/k8sclient"
	"goflow/orchestrator"
	"goflow/testutils"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var configPath string

func adjustConfigDagPath(configPath string) string {
	fixedConfig := &config.GoFlowConfig{}
	configBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(configBytes, fixedConfig)
	fixedConfig.DAGPath = testutils.GetDagsFolder()
	fmt.Println(fixedConfig)
	newConfigPath := filepath.Join(testutils.GetTestFolder(), "tmp_config.json")
	fixedConfig.SaveConfig(newConfigPath)
	return newConfigPath
}

func TestMain(m *testing.M) {
	configPath = adjustConfigDagPath(testutils.GetConfigPath())
	defer os.Remove(configPath)
	m.Run()
}

func TestStartServer(t *testing.T) {
	defer testutils.CleanUpJobs(k8sclient.CreateKubeClient())
	orch := orchestrator.NewOrchestrator(configPath)
	orch.Start(3)
	panic("test")
}
