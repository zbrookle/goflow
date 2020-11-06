package cron

import (
	"encoding/json"

	batch "k8s.io/api/batch/v1beta1"
)

func getJobJSONString(job batch.CronJob) string {
	hashString, _ := json.Marshal(job)
	return string(hashString)
}

func getJobFormattedJSONString(job batch.CronJob) string {
	jobJSON, _ := json.MarshalIndent(job, "", "\t")
	return string(jobJSON)
}
