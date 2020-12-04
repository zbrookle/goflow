package config

import (
	"goflow/config"
	"goflow/jsonpanic"
	"testing"

	core "k8s.io/api/core/v1"
)

func TestSetConfigDefaults(t *testing.T) {
	goflowConfig := config.GoFlowConfig{
		DefaultNamespace:     "default",
		DefaultDockerImage:   "busybox",
		DAGPath:              "path",
		DateFormat:           "2019-01-01",
		DefaultRestartPolicy: core.RestartPolicyNever,
		Parallelism:          1,
		TimeLimit:            1,
		Retries:              1,
		MaxActiveRuns:        1,
	}
	expectedDagConfig := DAGConfig{
		Name:        "test-config",
		Namespace:   goflowConfig.DefaultNamespace,
		Schedule:    "* * * * *",
		DockerImage: goflowConfig.DefaultDockerImage,
		RetryPolicy: core.RestartPolicyNever,
		Command: []string{
			"echo",
			"test",
		},
		Parallelism:   1,
		Retries:       1,
		MaxActiveRuns: 1,
		WithLogs:      false,
	}
	dagConfigCases := []DAGConfig{
		{
			Name:        expectedDagConfig.Name,
			Schedule:    expectedDagConfig.Schedule,
			RetryPolicy: core.RestartPolicyNever,
			Command:     []string{"echo", "test"},
			WithLogs:    false,
		},
		expectedDagConfig.Copy(),
	}
	for _, dagConfig := range dagConfigCases {
		dagConfig.SetDefaults(goflowConfig)
		configString := jsonpanic.JSONPanicFormat(dagConfig)
		expectedString := jsonpanic.JSONPanicFormat(expectedDagConfig)
		if configString != expectedString {
			t.Errorf("Expected\n%s\nbut found\n%s", expectedString, configString)
		}
	}
}
