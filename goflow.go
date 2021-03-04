package main

import (
	"flag"
	"goflow/internal/config"
	"goflow/internal/dag/metrics"
	"goflow/internal/dag/orchestrator"
	"goflow/internal/k8s/client"
	"goflow/internal/k8s/pod/utils"
	"goflow/internal/logs"
	"goflow/internal/paths"
	"goflow/internal/rest"
	"goflow/internal/termination"
	"goflow/internal/testutils"
	"io/ioutil"
	"time"

	core "k8s.io/api/core/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var host string
var port int

func main() {
	configPath := flag.String(
		"path",
		paths.GetGoDefaultHomePath(),
		"The path to the configuration file",
	)
	host := flag.String("host", "localhost", "Host IP to serve REST api on")
	port := flag.Int("port", 8080, "Port to serve REST API on")
	verbosePtr := flag.Bool("V", false, "Verbose logging")
	testMode := flag.Bool("T", false, "Uses test mode which leverage a mocked kubernetes client")
	flag.Parse()

	if !*verbosePtr {
		logs.InfoLogger.SetOutput(ioutil.Discard)
	}

	var orch *orchestrator.Orchestrator
	if *testMode {
		kubeClient := fake.NewSimpleClientset()
		kubeClient.Tracker().Add(&core.Namespace{
			ObjectMeta: v1.ObjectMeta{
				Name: "default",
			},
		})
		testutils.RegisterContainerStatusesToPods(kubeClient)
		config := config.CreateConfig(*configPath)
		config.DAGsOn = true
		orch = orchestrator.NewOrchestratorFromClientsAndConfig(
			kubeClient,
			config,
			metrics.NewDAGMetricsClient(kubeClient, true),
		)
	} else {
		defer utils.CleanUpEnvironment(client.CreateKubeClient())
		orch = orchestrator.NewOrchestrator(*configPath)
	}
	orch.Start(1 * time.Second)
	go termination.Handle(func() {
		orch.Stop()
	})
	go rest.Serve(*host, *port, orch)
	orch.Wait()
}
