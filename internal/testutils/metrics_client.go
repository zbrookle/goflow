package testutils

import (
	"goflow/internal/dag/metrics"

	fakemetrics "k8s.io/metrics/pkg/client/clientset/versioned/fake"
)

// NewTestMetricsClient returns a new metrics client for testing only
func NewTestMetricsClient() *metrics.DAGMetricsClient {
	fakeMetricsClientSet := fakemetrics.NewSimpleClientset()
	return metrics.NewDAGMetricsClient(fakeMetricsClientSet)
}
