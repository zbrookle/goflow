package config

import (
	"encoding/json"
	"goflow/internal/config"
	"goflow/internal/jsonpanic"
	"io/ioutil"
	"regexp"

	core "k8s.io/api/core/v1"
)

const validNameRegexString = "^[[:alpha:]][a-zA-Z0-9_-]+$"

var validNameRegex *regexp.Regexp

func init() {
	compiledRegex, err := regexp.Compile(validNameRegexString)
	if err != nil {
		panic(err)
	}
	validNameRegex = compiledRegex
}

// DAGConfig is a struct storing the configurable values provided from the user in the DAG
// definition file
type DAGConfig struct {
	Name          string
	Namespace     string
	Schedule      string
	DockerImage   string
	RetryPolicy   core.RestartPolicy
	Command       []string
	Parallelism   int32
	TimeLimit     *int64
	Retries       int32
	MaxActiveRuns int
	StartDateTime string
	EndDateTime   string
	Labels        map[string]string
	Annotations   map[string]string
	WithLogs      bool
}

// Marshal returns a json bytes representation of DAGConfig
func (config DAGConfig) Marshal() []byte {
	jsonBytes, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}
	return jsonBytes
}

// JSON returns a json string representation of DAGConfig
func (config DAGConfig) JSON() string {
	return string(config.Marshal())
}

func makeStrMapCopy(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	cpy := make(map[string]string)
	for key := range src {
		cpy[key] = src[key]
	}
	return cpy
}

// Copy returns a copy of the DAGConfig
func (config DAGConfig) Copy() DAGConfig {
	configCopy := config
	configCopy.Command = make([]string, len(config.Command))
	copy(configCopy.Command, config.Command)
	configCopy.Annotations = makeStrMapCopy(config.Annotations)
	configCopy.Labels = makeStrMapCopy(config.Labels)
	return configCopy
}

// SetDefaults sets the defaults on the dag config using the goflow config settings
func (config *DAGConfig) SetDefaults(goflowConfig config.GoFlowConfig) {
	if config.DockerImage == "" {
		config.DockerImage = goflowConfig.DefaultDockerImage
	}
	if config.Namespace == "" {
		config.Namespace = goflowConfig.DefaultNamespace
	}
	if config.RetryPolicy == "" {
		config.RetryPolicy = goflowConfig.DefaultRestartPolicy
	}
	if config.Parallelism == 0 {
		config.Parallelism = goflowConfig.Parallelism
	}
	if config.Retries == 0 {
		config.Retries = goflowConfig.Retries
	}
	if config.MaxActiveRuns == 0 {
		config.MaxActiveRuns = goflowConfig.MaxActiveRuns
	}
}

// IsNameValid returns false if name does not match required DAG naming pattern
func (config *DAGConfig) IsNameValid() bool {
	return validNameRegex.Match([]byte(config.Name))
}

// WriteToFile writes a dag file to the given folder
func (config *DAGConfig) WriteToFile(path string) error {
	return ioutil.WriteFile(path, jsonpanic.JSONPanicFormatBytes(config), 0600)
}

func (config *DAGConfig) String() string {
	return jsonpanic.JSONPanicFormat(config)
}

// Pattern returns the name pattern string that is used to determine valid names
func (config *DAGConfig) Pattern() string {
	return validNameRegexString
}
