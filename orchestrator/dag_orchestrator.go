package orchestrator

import (
	"context"
	"goflow/logs"

	"goflow/dags"

	"goflow/config"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// registerDag adds a job to the dictionary of DAGs
func (orchestrator Orchestrator) registerDag(dag *dags.DAG) {
	orchestrator.dagMap[dag.Name] = dag
}

// AddDAG adds a CronJob object to the Orchestrator and creates the job in kubernetes
func (orchestrator Orchestrator) AddDAG(dag *dags.DAG) {
	orchestrator.registerDag(dag)
}

// deleteDAG deletes a CronJob object in kubernetes
func (orchestrator Orchestrator) deleteDAG(jobName string, namespace string) {
	err := orchestrator.kubeClient.BatchV1beta1().CronJobs(
		namespace,
	).Delete(
		context.TODO(),
		jobName,
		v1.DeleteOptions{},
	)
	if err != nil {
		panic(err)
	}
	logs.InfoLogger.Printf("CronJob %s was deleted from namespace %s", jobName, namespace)
}

// RemoveDAG removes a CronJob object from the orchestrator
func (orchestrator Orchestrator) RemoveDAG(jobName string, namespace string) {
	orchestrator.deleteDAG(jobName, namespace)
	delete(orchestrator.dagMap, jobName)
}

// DAGs returns a CronJob list
func (orchestrator Orchestrator) DAGs() []*dags.DAG {
	jobs := make([]*dags.DAG, 0, len(orchestrator.dagMap))
	for job := range orchestrator.dagMap {
		jobs = append(jobs, orchestrator.dagMap[job])
	}
	return jobs
}

// CollectDAGs fills up the dag map with existing dags
func (orchestrator Orchestrator) CollectDAGs() {

}

// Start starts the orchestrator process
func (orchestrator Orchestrator) Start() {
	serverRunning := true
	for serverRunning {

	}
}
