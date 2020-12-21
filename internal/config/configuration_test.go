package config

import (
	"goflow/internal/jsonpanic"
	"goflow/internal/testutils"
	"testing"

	core "k8s.io/api/core/v1"
)

var configPath string

func TestMain(m *testing.M) {
	configPath = testutils.GetConfigPath()
	m.Run()
}

func TestReadConfig(t *testing.T) {
	foundConfig := CreateConfig(configPath)
	expectedConfig := GoFlowConfig{
		DefaultNamespace:     "default",
		DefaultDockerImage:   "busybox",
		DefaultRestartPolicy: core.RestartPolicyNever,
		Parallelism:          1,
		TimeLimit:            1,
		Retries:              1,
		MaxActiveRuns:        1,
		DAGPath:              "path",
		DateFormat:           "2019-01-01",
		DatabaseDNS:          "goflow.sqlite3",
	}
	if *foundConfig != expectedConfig {
		t.Error("Configs do not match")
		t.Errorf("Found: %s", jsonpanic.JSONPanicFormat(foundConfig))
		t.Errorf("Expected: %s", jsonpanic.JSONPanicFormat(expectedConfig))
	}
}
