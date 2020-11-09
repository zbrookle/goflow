package create_and_monitor_job

import (
	"goflow/orchestrator"
	testpaths "goflow/testutils"
	"testing"
)

func TestCreateAndMonitorJob(b *testing.B) {
	orch := orchestrator.NewOrchestrator(testpaths.GetConfigPath())
	orch.Start()
}
