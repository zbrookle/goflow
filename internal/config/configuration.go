package config

import (
	"encoding/json"
	"goflow/internal/jsonpanic"
	"goflow/internal/logs"
	"io/ioutil"

	core "k8s.io/api/core/v1"
)

// GoFlowConfig is a configuration struct for the GoFlow application settings
type GoFlowConfig struct {
	DefaultNamespace     string
	DefaultDockerImage   string
	DefaultRestartPolicy core.RestartPolicy
	Parallelism          int32
	TimeLimit            int64
	Retries              int32
	MaxActiveRuns        int
	DAGPath              string
	DateFormat           string
	DatabaseDNS          string
}

func readConfig(filePath string) []byte {
	dat, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	return dat
}

func verifyConfig(config GoFlowConfig) {
	if config.DefaultRestartPolicy == "" {
		panic("Restart policy must be specified!")
	}
	if config.DatabaseDNS == "" {
		panic("Database DNS must be specified!")
	}
}

// CreateConfig creates a configuration object based on the file at the given path
func CreateConfig(filePath string) *GoFlowConfig {
	configBytes := readConfig(filePath)
	configStruct := &GoFlowConfig{}
	err := json.Unmarshal(configBytes, configStruct)
	if err != nil {
		panic(err)
	}
	logs.InfoLogger.Println("Starting GoFlow with the following configs:")
	logs.InfoLogger.Println(jsonpanic.JSONPanicFormat(configStruct))
	verifyConfig(*configStruct)
	return configStruct
}

// SaveConfig saves the current in memory configuration to the config file
func (config GoFlowConfig) SaveConfig(filePath string) {
	configBytes, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(filePath, configBytes, 0666)
	if err != nil {
		panic(err)
	}
}
