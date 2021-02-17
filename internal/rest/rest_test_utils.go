package rest

import (
	"goflow/internal/config"
	"goflow/internal/dag/orchestrator"
	"goflow/internal/testutils"

	"k8s.io/client-go/kubernetes/fake"
)

// func createFakeKubeClient() *fake.Clientset {
// 	return fake.NewSimpleClientset()
// }

func getTestOrchestrator(configPath string) *orchestrator.Orchestrator {
	kubeClient := fake.NewSimpleClientset()
	configuration := config.CreateConfig(configPath)
	configuration.DAGPath = testutils.GetDagsFolder()
	configuration.DatabaseDNS = testutils.GetSQLiteLocation()
	return orchestrator.NewOrchestratorFromClientAndConfig(kubeClient, configuration)
}
