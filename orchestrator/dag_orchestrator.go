package orchestrator

import (
	"context"
	"goflow/logs"

	"goflow/dags"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Orchestrator holds information for all DAGs
type Orchestrator struct {
	dagMap    map[string]*dags.DAG
	kubeClient kubernetes.Interface
}

func newOrchestratorFromClient(client kubernetes.Interface) *Orchestrator {
	return &Orchestrator{make(map[string]*dags.DAG), client}
}

// NewOrchestrator creates an empty instance of Orchestrator
func NewOrchestrator() *Orchestrator {
	return newOrchestratorFromClient(createKubeClient())
}

// registerDag adds a job to the dictionary of DAGs
func (orchestrator Orchestrator) registerDag(dag *dags.DAG) {
	orchestrator.dagMap[dag.Name] = dag
}

// createKubeJob creates a k8s cron job in k8s
// func (orchestrator Orchestrator) createKubeJob(job *batch.CronJob) *batch.CronJob {
// 	createdJob, err := orchestrator.kubeClient.BatchV1beta1().CronJobs(
// 		"default",
// 	).Create(
// 		context.TODO(),
// 		job,
// 		v1.CreateOptions{},
// 	)
// 	if err != nil {
// 		panic(err)
// 	}

// 	logs.InfoLogger.Printf(
// 		"Created CronJob %q.\n, with configuration %s",
// 		createdJob.GetObjectMeta().GetName(),
// 		cron.GetJobFormattedJSONString(*createdJob),
// 	)
// 	return createdJob
// }

// AddDAG adds a CronJob object to the Orchestrator and creates the job in kubernetes
func (orchestrator Orchestrator) AddDAG(job *dags.DAG) {
	createdDAG := dags
	orchestrator.registerDag(createdDAG)
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

// AddNewDAGs fills up the jobs layer with existing dags
func (orchestrator Orchestrator) AddNewDAGs() {

}

// Start starts the orchestrator process
func (orchestrator Orchestrator) Start() {
	serverRunning := true
	for serverRunning {
		
	}
}