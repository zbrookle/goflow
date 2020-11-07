package cron

import (
	"encoding/json"

	batch "k8s.io/api/batch/v1beta1"
)

func getJobJSONString(job batch.CronJob) string {
	hashString, _ := json.Marshal(job)
	return string(hashString)
}

// GetJobFormattedJSONString returns a JSON formatted string of a CronJob
func GetJobFormattedJSONString(job batch.CronJob) string {
	jobJSON, _ := json.MarshalIndent(job, "", "\t")
	return string(jobJSON)
}
