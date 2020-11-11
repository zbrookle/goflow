package dags

import "encoding/json"

// DAGConfig is a struct storing the configurable values provided from the user in the DAG
// definition file
type DAGConfig struct {
	Name          string
	Namespace     string
	Schedule      string
	DockerImage   string
	RetryPolicy   string
	Command       string
	Parallelism   int32
	TimeLimit     int64
	Retries       int32
	MaxActiveRuns int
	StartDateTime string
	EndDateTime   string
	Labels        map[string]string
	Annotations   map[string]string
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
