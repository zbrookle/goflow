package config

import (
	"goflow/podutils"
	"testing"
)

var configPath string

func TestMain(m *testing.M) {
	configPath = podutils.GetConfigPath()
	m.Run()
}

func TestReadConfig(t *testing.T) {
	foundConfig := CreateConfig(configPath)
	expectedConfig := GoFlowConfig{"default", "busybox", "path", "2019-01-01"}
	if *foundConfig != expectedConfig {
		t.Error("Configs do not match")
		t.Errorf("Found: %s", foundConfig)
		t.Errorf("Expected: %s", expectedConfig)
	}
}
