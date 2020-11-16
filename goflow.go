package main

import (
	"flag"
	"goflow/k8sclient"
	"goflow/logs"
	"goflow/orchestrator"
	"goflow/testutils"
	"time"
)

func main() {
	configPath := flag.String(
		"path",
		testutils.GetConfigPath(),
		"The path to the configuration file",
	)
	flag.Parse()

	kubeClient := k8sclient.CreateKubeClient()

	defer testutils.CleanUpPods(kubeClient)
	orch := *orchestrator.NewOrchestrator(*configPath)
	loopBreaker := false
	go orch.Start(1, &loopBreaker)

	time.Sleep(4 * time.Second)
	loopBreaker = true

	logs.InfoLogger.Println("Dags length", len(orch.DAGs()))
}
