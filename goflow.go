package main

import (
	"flag"
	"goflow/internal/dag/orchestrator"
	"goflow/internal/k8s/client"
	"goflow/internal/k8s/pod/utils"
	"goflow/internal/paths"
	"goflow/internal/rest"
	"goflow/internal/termination"
	"time"
)

var host string
var port int


func main() {
	defer utils.CleanUpEnvironment(client.CreateKubeClient())
	configPath := flag.String(
		"path",
		paths.GetGoDefaultHomePath(),
		"The path to the configuration file",
	)
	host := flag.String("host", "localhost", "Host IP to serve REST api on")
	port := flag.Int("port", 8080, "Port to serve REST API on")
	flag.Parse()

	orch := orchestrator.NewOrchestrator(*configPath)
	orch.Start(1 * time.Second)
	go termination.Handle(func() {
		orch.Stop()
	})
	go rest.Serve(*host, *port, orch)
	orch.Wait()
}
