package main

import (
	"flag"
	"goflow/dag/orchestrator"
	"goflow/paths"
	"time"
)

func main() {
	configPath := flag.String(
		"path",
		paths.GetGoDefaultHomePath(),
		"The path to the configuration file",
	)
	flag.Parse()

	orch := *orchestrator.NewOrchestrator(*configPath)
	loopBreaker := make(chan struct{}, 1)
	orch.Start(1*time.Second, loopBreaker)

	for {

	}
}
