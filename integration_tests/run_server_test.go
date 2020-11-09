package servertest

import (
	"encoding/json"
	"goflow/config"
	"goflow/k8sclient"
	"goflow/orchestrator"
	"goflow/testutils"
	"io/ioutil"
	"testing"
)

var configPath string

func adjustConfigDagPath(configPath string) {
	fixedConfig := config.GoFlowConfig{}
	configBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(configBytes, fixedConfig)
	fixedConfig.DAGPath = testutils.GetDagsFolder()
	fixedConfig.SaveConfig(configPath)
}

func TestMain(m *testing.M) {
	configPath = testutils.GetConfigPath()
	adjustConfigDagPath(configPath)
	m.Run()
}

func BenchmarkStartServer(b *testing.B) {
	defer testutils.CleanUpJobs(k8sclient.CreateKubeClient())
	orch := orchestrator.NewOrchestrator(configPath)
	orch.Start(1)
}
