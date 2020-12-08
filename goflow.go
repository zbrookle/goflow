package main

import (
	"flag"
	"goflow/dag/orchestrator"
	"goflow/k8s/client"
	"goflow/k8s/pod/utils"
	"goflow/paths"
	"goflow/termination"
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
