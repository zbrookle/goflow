package dags

import (
	"context"
	"goflow/jsonpanic"
	"goflow/testutils"
	"testing"
	"time"

	batch "k8s.io/api/batch/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateJob(t *testing.T) {
	defer testutils.CleanUpJobs(KUBECLIENT)
	dagRun := createDagRun(getTestDate(), getTestDAGFakeClient())
	dagRun.createJob()
	foundJob, err := dagRun.DAG.kubeClient.BatchV1().Jobs(
		dagRun.DAG.Config.Namespace,
	).Get(
		context.TODO(),
		dagRun.job.Name,
		v1.GetOptions{},
	)
	if err != nil {
		panic(err)
	}
	foundJobValue := jsonpanic.JSONPanic(*foundJob)
	expectedValue := jsonpanic.JSONPanic(*dagRun.job)
	if foundJobValue != expectedValue {
		t.Error("Expected:", expectedValue)
		t.Error("Found:", foundJobValue)
	}
}

func TestStartJob(t *testing.T) {
	defer testutils.CleanUpJobs(KUBECLIENT)
	dagRun := createDagRun(getTestDate(), getTestDAGRealClient())
	jobNameChannel := make(chan string, 1)
	dagRun.Start(jobNameChannel)
	jobName := <-jobNameChannel
	job, err := KUBECLIENT.BatchV1().Jobs(
		dagRun.DAG.Config.Namespace,
	).Get(
		context.TODO(),
		jobName,
		v1.GetOptions{},
	)

	time.Sleep(6 * time.Second)

	if err != nil {
		t.Errorf("Job %s could not be found", jobName)
	}

	if job.Status.Succeeded != 1 {
		t.Errorf("Job %s did not complete yet", jobName)
	}
}

func unmarshalJob(job batch.Job) string {
	bytes := make([]byte, 0)
	err := job.Unmarshal(bytes)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func TestDeleteJob(t *testing.T) {
	defer testutils.CleanUpJobs(KUBECLIENT)
	dagRun := createDagRun(getTestDate(), getTestDAGFakeClient())
	jobFrame := dagRun.getJobFrame()
	jobsClient := KUBECLIENT.BatchV1().Jobs(dagRun.DAG.Config.Namespace)

	createdJob, err := jobsClient.Create(context.TODO(), &jobFrame, v1.CreateOptions{})
	dagRun.job = createdJob
	if err != nil {
		panic(err)
	}
	dagRun.deleteJob()
	list, err := jobsClient.List(context.TODO(), v1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, job := range list.Items {
		if unmarshalJob(*createdJob) == unmarshalJob(job) {
			t.Errorf("Job %s should have been deleted", createdJob.Name)
		}
	}
}
