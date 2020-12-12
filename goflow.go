package main

import (
	"flag"
	"goflow/internal/dag/orchestrator"
	"goflow/internal/k8s/client"
	"goflow/internal/k8s/pod/utils"
	"goflow/internal/paths"
	"goflow/internal/termination"
	"time"
)

func main() {
	defer utils.CleanUpEnvironment(client.CreateKubeClient())
	configPath := flag.String(
		"path",
		paths.GetGoDefaultHomePath(),
		"The path to the configuration file",
	)
	flag.Parse()

	orch := *orchestrator.NewOrchestrator(*configPath)
	orch.Start(1 * time.Second)
	go termination.Handle(func() {
		orch.Stop()
	})
	orch.Wait()
}
