package cron

import (
	"context"

	batch "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Orchestrator holds information for all cronjobs
type Orchestrator struct {
	cronMap map[string]*batch.CronJob
	kubeClient *kubernetes.Clientset
}

// NewOrchestrator creates an empty instance of Orchestrator
func NewOrchestrator() *Orchestrator {
	return &Orchestrator{make(map[string]*batch.CronJob), CreateKubeClient()}
}

// AddJob adds a CronJob object to the Orchestrator
func (orchestrator Orchestrator) AddJob(job *batch.CronJob) {
	orchestrator.cronMap[job.ObjectMeta.Name] = job
	orchestrator.kubeClient.BatchV1beta1().CronJobs("default").Create(context.TODO(), job, v1.CreateOptions{})
}

// Jobs returns a CronJob list
func (orchestrator Orchestrator) Jobs() []*batch.CronJob {
	jobs := make([]*batch.CronJob, 0, len(orchestrator.cronMap))
	for job := range orchestrator.cronMap {
		jobs = append(jobs, orchestrator.cronMap[job])
	}
	return jobs
}

// fillJobs fills up the jobs layer with existing scheduled kubernetes jobs
func (orchestrator Orchestrator) fillJobs() {

}
