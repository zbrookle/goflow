package cron

import (
	batch "k8s.io/api/batch/v1beta1"
	batchclient "k8s.io/client-go/kubernetes/typed/batch/v1"
)

// Orchestrator holds information for all cronjobs
type Orchestrator struct {
	cronMap map[string]*batch.CronJob
	batchClient *batchclient.BatchV1Client
}

// NewOrchestrator creates an empty instance of Orchestrator
func NewOrchestrator() *Orchestrator {
	return &Orchestrator{make(map[string]*batch.CronJob), CreateKubeBatchClient()}
}

// AddJob adds a CronJob object to the Orchestrator
func (orchestrator Orchestrator) AddJob(job *batch.CronJob) {
	orchestrator.cronMap[job.ObjectMeta.Name] = job
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
