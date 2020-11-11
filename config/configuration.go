package config

import (
	"encoding/json"
	"io/ioutil"
)

// GoFlowConfig is a configuration struct for the GoFlow application settings
type GoFlowConfig struct {
	DefaultNamespace   string
	DefaultDockerImage string
	DAGPath            string
	DateFormat         string
}

func readConfig(filePath string) []byte {
	dat, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	return dat
}

// CreateConfig creates a configuration object based on the file at the given path
func CreateConfig(filePath string) *GoFlowConfig {
	configBytes := readConfig(filePath)
	emptyConfig := &GoFlowConfig{}
	err := json.Unmarshal(configBytes, emptyConfig)
	if err != nil {
		panic(err)
	}
	return emptyConfig
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
