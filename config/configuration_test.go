package config

import (
	"goflow/testpaths"
	"testing"
)

var configPath string

func TestMain(m *testing.M){
	configPath = testpaths.GetConfigPath()
	m.Run()
}

func TestReadConfig(t *testing.T) {
	foundConfig := CreateConfig(configPath)
	expectedConfig := GoFlowConfig{"default", "busybox"}
	if *foundConfig != expectedConfig {
		t.Error("Configs do not match")
		t.Errorf("Found: %s", foundConfig)
		t.Errorf("Expected: %s", expectedConfig)
	}
}