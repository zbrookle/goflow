package cron

import (
	"fmt"
	batch "k8s.io/api/batch/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func CreateCronJob(cronID int) *batch.CronJob {
	var cronName = fmt.Sprint("cronJob&d", cronID)
	var kubeType = meta.TypeMeta{Kind: "CronJob", APIVersion: batch.SchemeGroupVersion.Version}
	var objectMeta = meta.ObjectMeta{Name: cronName}
	var cronSpec = batch.CronJobSpec{Schedule: "* * * * *", JobTemplate: batch.JobTemplateSpec{}}
	return &batch.CronJob{TypeMeta: kubeType, ObjectMeta: objectMeta, Spec: cronSpec, Status: batch.CronJobStatus{}}
}

func TestAddCronJobToOrchestrator(t *testing.T) {
	var job = CreateCronJob(0)
	var orch = NewOrchestrator()
	const expectedLength = 1
	orch.AddJob(job)
	if orch.cronMap[job.ObjectMeta.Name] != job {
		t.Error("CronJob not added at correct key")
	}
	if len(orch.cronMap) != expectedLength {
		t.Errorf("CronMap should have length %d", expectedLength)
	}
}

func TestGetCronJobs(t *testing.T) {
	var cronJobs = []*batch.CronJob{CreateCronJob(0), CreateCronJob(1)}
	var cronMap = make(map[string]*batch.CronJob)
	for _, job := range cronJobs {
		cronMap[job.ObjectMeta.Name] = job
	}
	var orch = Orchestrator{cronMap: cronMap}
	for j, job := range orch.Jobs() {
		if cronJobs[j] != job {
			t.Error("CronJob lists do not match")
		}
	}
}
