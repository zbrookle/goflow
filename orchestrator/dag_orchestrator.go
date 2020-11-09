package orchestrator

import (
	"fmt"
	"goflow/dags"

	"goflow/config"

	"k8s.io/client-go/kubernetes"
)

// Orchestrator holds information for all DAGs
type Orchestrator struct {
	dagMap     map[string]*dags.DAG
	kubeClient kubernetes.Interface
	config     *config.GoFlowConfig
}

func newOrchestratorFromClientAndConfig(
	client kubernetes.Interface,
	config *config.GoFlowConfig,
) *Orchestrator {
	return &Orchestrator{make(map[string]*dags.DAG), client, config}
}

// NewOrchestrator creates an empty instance of Orchestrator
func NewOrchestrator(configPath string) *Orchestrator {
	return newOrchestratorFromClientAndConfig(createKubeClient(), config.CreateConfig(configPath))
}

// AddDAG adds a CronJob object to the Orchestrator and creates the job in kubernetes
func (orchestrator *Orchestrator) AddDAG(dag *dags.DAG) {
	orchestrator.dagMap[dag.Name] = dag
}

// DeleteDAG removes a CronJob object from the orchestrator
func (orchestrator Orchestrator) DeleteDAG(dagName string, namespace string) {
	// dag := orchestrator.dagMap[dagName]
	// dag.TerminateAndDeleteJobs()
	delete(orchestrator.dagMap, dagName)
}

// DAGs returns []DAGs with all DAGs present in the map
func (orchestrator Orchestrator) DAGs() []*dags.DAG {
	jobs := make([]*dags.DAG, 0, len(orchestrator.dagMap))
	for job := range orchestrator.dagMap {
		jobs = append(jobs, orchestrator.dagMap[job])
	}
	return jobs
}

// CollectDAGs fills up the dag map with existing dags
func (orchestrator *Orchestrator) CollectDAGs() {
	dagSlice := dags.GetDAGSFromFolder(orchestrator.config.DAGPath)
	fmt.Print(dagSlice)
	for _, dag := range dagSlice {
		fmt.Print(dag)
		orchestrator.AddDAG(&dag)
	}
}

// Start starts the orchestrator process
func (orchestrator Orchestrator) Start() {
	serverRunning := true
	for serverRunning {

	}
}
