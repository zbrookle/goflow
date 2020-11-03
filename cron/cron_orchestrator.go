package cron

import batchv1 "k8s.io/api/batch/v1beta1"

// Orchestrator holds information for all cronjobs
type Orchestrator struct {
	cronMap map[string]batchv1.CronJob
}

func addJob(orchestrator Orchestrator, job batchv1.CronJob) {
	orchestrator.cronMap[job.ObjectMeta.Name] = job
}