package orchestrator

import (
	"context"
	"fmt"
	"testing"

	"goflow/cron"

	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var kubeClient *kubernetes.Clientset

func TestMain(m *testing.M) {
	kubeClient = createKubeClient()
	m.Run()
}

func CreateCronJob(cronID int) *batchv1beta1.CronJob {
	var cronName = fmt.Sprintf("cron-job-%d", cronID)
	var kubeType = meta.TypeMeta{Kind: "CronJob", APIVersion: batchv1beta1.SchemeGroupVersion.Version}
	var objectMeta = meta.ObjectMeta{Name: cronName}
	var cronSpec = batchv1beta1.CronJobSpec{Schedule: "* * * * *", JobTemplate: batchv1beta1.JobTemplateSpec{Spec: batchv1.JobSpec{Template: core.PodTemplateSpec{
		ObjectMeta: meta.ObjectMeta{
			Labels: map[string]string{
				"app": "demo",
			},
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:  "web",
					Image: "nginx:1.12",
					Ports: []core.ContainerPort{
						{
							Name:          "http",
							Protocol:      core.ProtocolTCP,
							ContainerPort: 80,
						},
					},
				},
			},
			RestartPolicy: "Never",
		},
	}}}}
	return &batchv1beta1.CronJob{
		TypeMeta:   kubeType,
		ObjectMeta: objectMeta,
		Spec:       cronSpec,
		Status:     batchv1beta1.CronJobStatus{},
	}
}

func TestRegisterCronJob(t *testing.T) {
	job := CreateCronJob(0)
	orch := NewOrchestrator()
	const expectedLength = 1
	orch.registerJob(job)
	if orch.cronMap[job.ObjectMeta.Name] != job {
		t.Error("CronJob not added at correct key")
	}
	if len(orch.cronMap) != expectedLength {
		t.Errorf("CronMap should have length %d", expectedLength)
	}
}

func TestCreateCronJobInK8S(t *testing.T) {
	job := CreateCronJob(0)
	orch := NewOrchestrator()
	createdJob := orch.createKubeJob(job)

	expectedJobsSet := cron.NewSetFromList([]batchv1beta1.CronJob{*createdJob})
	namespace := "default"

	// Retrieve jobs that are present in k8s
	retrievedJobListObject, err := kubeClient.BatchV1beta1().CronJobs(
		namespace,
	).List(
		context.TODO(),
		meta.ListOptions{},
	)
	if err != nil {
		panic(err)
	}

	// Delete test job
	kubeClient.BatchV1beta1().CronJobs(
		namespace,
	).Delete(
		context.TODO(),
		job.Name,
		meta.DeleteOptions{},
	)

	retrievedJobSet := cron.NewSetFromList(retrievedJobListObject.Items)
	if !retrievedJobSet.Equals(&expectedJobsSet) {
		t.Errorf(
			"Retrieved job list did not match expected job list. \nExpected: %s\nRetrieved:%s",
			expectedJobsSet.String(),
			retrievedJobSet.String(),
		)
	}
}

func TestGetCronJobs(t *testing.T) {
	var cronJobs = []*batchv1beta1.CronJob{CreateCronJob(0), CreateCronJob(1)}
	var cronMap = make(map[string]*batchv1beta1.CronJob)
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
