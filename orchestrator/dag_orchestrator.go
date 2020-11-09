package orchestrator

import (
	"context"
	"goflow/logs"

	"goflow/cron"
	"goflow/dags"

	batch "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Orchestrator holds information for all cronjobs
type Orchestrator struct {
	cronMap    map[string]*batch.CronJob
	kubeClient kubernetes.Interface
}

func newOrchestratorFromClient(client kubernetes.Interface) *Orchestrator {
	return &Orchestrator{make(map[string]*batch.CronJob), client}
}

// NewOrchestrator creates an empty instance of Orchestrator
func NewOrchestrator() *Orchestrator {
	return newOrchestratorFromClient(createKubeClient())
}

// registerDag adds a job to the dictionary of jobs
func (orchestrator Orchestrator) registerDag(dag *dags.DAG) {
	orchestrator.cronMap[job.ObjectMeta.Name] = job
}

// createKubeJob creates a k8s cron job in k8s
func (orchestrator Orchestrator) createKubeJob(job *batch.CronJob) *batch.CronJob {
	createdJob, err := orchestrator.kubeClient.BatchV1beta1().CronJobs(
		"default",
	).Create(
		context.TODO(),
		job,
		v1.CreateOptions{},
	)
	if err != nil {
		panic(err)
	}

	logs.InfoLogger.Printf(
		"Created CronJob %q.\n, with configuration %s",
		createdJob.GetObjectMeta().GetName(),
		cron.GetJobFormattedJSONString(*createdJob),
	)
	return createdJob
}

// AddJob adds a CronJob object to the Orchestrator and creates the job in kubernetes
func (orchestrator Orchestrator) AddJob(job *batch.CronJob) {
	createdJob := orchestrator.createKubeJob(job)
	orchestrator.registerDag(createdJob)
}

// deleteKubeJob deletes a CronJob object in kubernetes
func (orchestrator Orchestrator) deleteKubeJob(jobName string, namespace string) {
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

// RemoveJob removes a CronJob object from the orchestrator
func (orchestrator Orchestrator) RemoveJob(jobName string, namespace string) {
	orchestrator.deleteKubeJob(jobName, namespace)
	delete(orchestrator.cronMap, jobName)
}

// Jobs returns a CronJob list
func (orchestrator Orchestrator) Jobs() []*batch.CronJob {
	jobs := make([]*batch.CronJob, 0, len(orchestrator.cronMap))
	for job := range orchestrator.cronMap {
		jobs = append(jobs, orchestrator.cronMap[job])
	}
	return jobs
}

// AddNewJobs fills up the jobs layer with existing dags
func (orchestrator Orchestrator) AddNewJobs() {

}

// Start starts the orchestrator process
func (orchestrator Orchestrator) Start() {
	serverRunning := true
	for serverRunning {
		
	}
}