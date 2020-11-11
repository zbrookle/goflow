package dags

import (
	"context"

	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DAGRun is a single run of a given dag - corresponds with a kubernetes Job
type DAGRun struct {
	Name          string
	DAG           *DAG
	ExecutionDate k8sapi.Time // This is the date that will be passed to the job that runs
	Start         k8sapi.Time
	End           k8sapi.Time
	Job           *batch.Job
}

// getJobFrame returns a job from a DagRun
func (dagRun DAGRun) getJobFrame() batch.Job {
	dag := dagRun.DAG
	return batch.Job{
		TypeMeta: k8sapi.TypeMeta{
			Kind:       "Job",
			APIVersion: "v1",
		},
		ObjectMeta: k8sapi.ObjectMeta{
			Name:        dagRun.Name,
			Namespace:   dag.Config.Namespace,
			Labels:      dag.Config.Labels,
			Annotations: dag.Config.Annotations,
		},
		Spec: batch.JobSpec{
			Parallelism:           &dag.Config.Parallelism,
			ActiveDeadlineSeconds: &dag.Config.TimeLimit,
			BackoffLimit:          &dag.Config.Retries,
			Template: core.PodTemplateSpec{
				ObjectMeta: k8sapi.ObjectMeta{
					Name:      dag.Config.Name,
					Namespace: dag.Config.Namespace,
					// Labels: map[string]string{
					// 	"": "",
					// },
					// Annotations: map[string]string{
					// 	"": "",
					// },
				},
				Spec: core.PodSpec{
					Volumes:                       nil,
					Containers:                    nil,
					EphemeralContainers:           nil,
					RestartPolicy:                 "",
					TerminationGracePeriodSeconds: nil,
					ActiveDeadlineSeconds:         nil,
				},
			},
		},
	}
}

// CreateJob creates and registers a new job with
func (dagRun *DAGRun) CreateJob() {
	dag := dagRun.DAG
	jobFrame := dagRun.getJobFrame()
	job, err := dag.kubeClient.BatchV1().Jobs(
		dag.Config.Namespace,
	).Create(
		context.TODO(),
		&jobFrame,
		k8sapi.CreateOptions{},
	)
	if err != nil {
		panic(err)
	}
	dagRun.Job = job
}

// deleteJob
func (dagRun *DAGRun) deleteJob() {
	err := dagRun.DAG.kubeClient.BatchV1().Jobs(
		dagRun.DAG.Config.Namespace,
	).Delete(
		context.TODO(),
		dagRun.Name,
		k8sapi.DeleteOptions{},
	)
	if err != nil {
		panic(err)
	}
}
