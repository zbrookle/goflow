package main

import (
	"goflow/cron"
	// "encoding/json"
	// "github.com/davecgh/go-spew/spew"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

func main() {
	// spew.Config.Indent = "	"
	var orchestrator = cron.NewOrchestrator()
	job := CreateCronJob(0)
	orchestrator.AddJob(job)
	// job := orchestrator.Jobs()[0]
	// // spew.Dump(job)
	// stringRep, err := json.MarshalIndent(job, "", "\t")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("%s\n", stringRep)
	orchestrator.RemoveJob(job.Name, "default")
}
