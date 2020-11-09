package servertest

import (
	"goflow/orchestrator"
	testpaths "goflow/testutils"
	"testing"
)

func BenchmarkStartServer(b *testing.B) {
	orch := orchestrator.NewOrchestrator(testpaths.GetConfigPath())
	orch.Start()
}
