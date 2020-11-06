package cron

import (
	"encoding/json"
	"testing"

	batch "k8s.io/api/batch/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func stringInMap(s string, sMap map[string]bool) bool {
	_, ok := sMap[s]
	return ok
}

func TestCronJobAdd(t *testing.T) {
	jobSet := NewSet()
	job := batch.CronJob{ObjectMeta: meta.ObjectMeta{Name: "test"}}
	jobJSON, _ := json.Marshal(job)
	jobString := string(jobJSON)
	if stringInMap(jobString, jobSet.cronJobMap) {
		t.Errorf("Job %s should not be in cronjob map", jobString)
	}

	jobSet.Add(job)

	if !stringInMap(jobString, jobSet.cronJobMap) {
		t.Errorf("Job %s should be in cronjob map", jobString)
	}
}
