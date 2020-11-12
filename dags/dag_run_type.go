package dags

import (
	"context"
	"fmt"

	"strings"

	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// DAGRun is a single run of a given dag - corresponds with a kubernetes Job
type DAGRun struct {
	Name          string
	DAG           *DAG
	ExecutionDate k8sapi.Time // This is the date that will be passed to the job that runs
	StartTime     k8sapi.Time
	EndTime       k8sapi.Time
	job           *batch.Job
}

func (dagRun *DAGRun) getContainerFrame() core.Container {
	return core.Container{
		Name:       "job",
		Image:      dagRun.DAG.Config.DockerImage,
		Command:    strings.Split(dagRun.DAG.Config.Command, " "),
		Args:       nil,
		WorkingDir: "",
		EnvFrom:    nil,
		Env:        nil,
		// Resources: core.ResourceRequirements{
		// 	Limits: map[core.ResourceName]resource.Quantity{
		// 		"": {
		// 			Format: "",
		// 		},
		// 	},
		// 	Requests: map[core.ResourceName]resource.Quantity{
		// 		"": {
		// 			Format: "",
		// 		},
		// 	},
		// },
		VolumeMounts:  nil,
		VolumeDevices: nil,
		// LivenessProbe: &core.Probe{
		// 	Handler: core.Handler{
		// 		Exec: &core.ExecAction{
		// 			Command: nil,
		// 		},
		// 		HTTPGet: &core.HTTPGetAction{
		// 			Path: "",
		// 			Port: intstr.IntOrString{
		// 				Type:   0,
		// 				IntVal: 0,
		// 				StrVal: "",
		// 			},
		// 			Host:        "",
		// 			Scheme:      "",
		// 			HTTPHeaders: nil,
		// 		},
		// 		TCPSocket: &core.TCPSocketAction{
		// 			Port: intstr.IntOrString{
		// 				Type:   0,
		// 				IntVal: 0,
		// 				StrVal: "",
		// 			},
		// 			Host: "",
		// 		},
		// 	},
		// 	InitialDelaySeconds: 0,
		// 	TimeoutSeconds:      0,
		// 	PeriodSeconds:       0,
		// 	SuccessThreshold:    0,
		// 	FailureThreshold:    0,
		// },
		// ReadinessProbe: &core.Probe{
		// 	Handler: core.Handler{
		// 		Exec: &core.ExecAction{
		// 			Command: nil,
		// 		},
		// 		HTTPGet: &core.HTTPGetAction{
		// 			Path: "",
		// 			Port: intstr.IntOrString{
		// 				Type:   0,
		// 				IntVal: 0,
		// 				StrVal: "",
		// 			},
		// 			Host:        "",
		// 			Scheme:      "",
		// 			HTTPHeaders: nil,
		// 		},
		// 		TCPSocket: &core.TCPSocketAction{
		// 			Port: intstr.IntOrString{
		// 				Type:   0,
		// 				IntVal: 0,
		// 				StrVal: "",
		// 			},
		// 			Host: "",
		// 		},
		// 	},
		// 	InitialDelaySeconds: 0,
		// 	TimeoutSeconds:      0,
		// 	PeriodSeconds:       0,
		// 	SuccessThreshold:    0,
		// 	FailureThreshold:    0,
		// },
		// StartupProbe: &core.Probe{
		// 	Handler: core.Handler{
		// 		Exec: &core.ExecAction{
		// 			Command: nil,
		// 		},
		// 		HTTPGet: &core.HTTPGetAction{
		// 			Path: "",
		// 			Port: intstr.IntOrString{
		// 				Type:   0,
		// 				IntVal: 0,
		// 				StrVal: "",
		// 			},
		// 			Host:        "",
		// 			Scheme:      "",
		// 			HTTPHeaders: nil,
		// 		},
		// 		TCPSocket: &core.TCPSocketAction{
		// 			Port: intstr.IntOrString{
		// 				Type:   0,
		// 				IntVal: 0,
		// 				StrVal: "",
		// 			},
		// 			Host: "",
		// 		},
		// 	},
		// 	InitialDelaySeconds: 0,
		// 	TimeoutSeconds:      0,
		// 	PeriodSeconds:       0,
		// 	SuccessThreshold:    0,
		// 	FailureThreshold:    0,
		// },
		// Lifecycle: &core.Lifecycle{
		// 	PostStart: &core.Handler{
		// 		Exec: &core.ExecAction{
		// 			Command: nil,
		// 		},
		// 		HTTPGet: &core.HTTPGetAction{
		// 			Path: "",
		// 			Port: intstr.IntOrString{
		// 				Type:   0,
		// 				IntVal: 0,
		// 				StrVal: "",
		// 			},
		// 			Host:        "",
		// 			Scheme:      "",
		// 			HTTPHeaders: nil,
		// 		},
		// 		TCPSocket: &core.TCPSocketAction{
		// 			Port: intstr.IntOrString{
		// 				Type:   0,
		// 				IntVal: 0,
		// 				StrVal: "",
		// 			},
		// 			Host: "",
		// 		},
		// 	},
		// 	PreStop: &core.Handler{
		// 		Exec: &core.ExecAction{
		// 			Command: nil,
		// 		},
		// 		HTTPGet: &core.HTTPGetAction{
		// 			Path: "",
		// 			Port: intstr.IntOrString{
		// 				Type:   0,
		// 				IntVal: 0,
		// 				StrVal: "",
		// 			},
		// 			Host:        "",
		// 			Scheme:      "",
		// 			HTTPHeaders: nil,
		// 		},
		// 		TCPSocket: &core.TCPSocketAction{
		// 			Port: intstr.IntOrString{
		// 				Type:   0,
		// 				IntVal: 0,
		// 				StrVal: "",
		// 			},
		// 			Host: "",
		// 		},
		// 	},
		// },
		// TerminationMessagePath:   "",
		// TerminationMessagePolicy: "",
		ImagePullPolicy: "IfNotPresent",
		// SecurityContext: &core.SecurityContext{
		// 	Capabilities: &core.Capabilities{
		// 		Add:  nil,
		// 		Drop: nil,
		// 	},
		// 	Privileged: nil,
		// 	SELinuxOptions: &core.SELinuxOptions{
		// 		User:  "",
		// 		Role:  "",
		// 		Type:  "",
		// 		Level: "",
		// 	},
		// 	WindowsOptions: &core.WindowsSecurityContextOptions{
		// 		GMSACredentialSpecName: nil,
		// 		GMSACredentialSpec:     nil,
		// 		RunAsUserName:          nil,
		// 	},
		// 	RunAsUser:                nil,
		// 	RunAsGroup:               nil,
		// 	RunAsNonRoot:             nil,
		// 	ReadOnlyRootFilesystem:   nil,
		// 	AllowPrivilegeEscalation: nil,
		// 	ProcMount:                nil,
		// 	SeccompProfile: &core.SeccompProfile{
		// 		Type:             "",
		// 		LocalhostProfile: nil,
		// 	},
		// },
		// Stdin:     false,
		// StdinOnce: false,
		// TTY:       false,
	}
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
					Containers:                    []core.Container{dagRun.getContainerFrame()},
					EphemeralContainers:           nil,
					RestartPolicy:                 dag.Config.RetryPolicy,
					TerminationGracePeriodSeconds: nil,
					ActiveDeadlineSeconds:         nil,
				},
			},
		},
	}
}

// createJob creates and registers a new job with
func (dagRun *DAGRun) createJob() string {
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
	dagRun.job = job
	return job.Name
}

func (dagRun *DAGRun) monitorJob() watch.Event {
	podClient := dagRun.DAG.kubeClient.CoreV1().Pods(
		dagRun.job.Namespace,
	)

	watcher, err := podClient.Watch(context.TODO(), k8sapi.ListOptions{})
	if err != nil {
		panic(err)
	}
	result := <-watcher.ResultChan()
	panic(result)
	fmt.Println("Result", result)
	return result
	// req := podClient.GetLogs(
	// 	dagRun.job.Name,
	// 	&core.PodLogOptions{},
	// )
	// logs, err := req.Stream(context.TODO())
	// if err != nil {
	// 	panic(err)
	// }
	// dagRun.DAG.kubeClient.CoreV1().Pods(dagRun.job.Namespace).Watch()
	// fmt.Println("Logs", logs.)
	// return logs.Read()
}

// Start starts and monitors the job and also tracks the logs from the job
func (dagRun *DAGRun) Start(jobChannel chan string) {
	jobChannel <- dagRun.createJob()
	dagRun.monitorJob()
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
