package dags

import (
	"context"

	"strings"

	core "k8s.io/api/core/v1"
	k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// DAGRun is a single run of a given dag - corresponds with a kubernetes pod
type DAGRun struct {
	Name          string
	DAG           *DAG
	ExecutionDate k8sapi.Time // This is the date that will be passed to the pod that runs
	StartTime     k8sapi.Time
	EndTime       k8sapi.Time
	pod           *core.Pod
}

func (dagRun *DAGRun) getContainerFrame() core.Container {
	command := strings.Split(dagRun.DAG.Config.Command, " ")
	return core.Container{
		Name:       "task",
		Image:      dagRun.DAG.Config.DockerImage,
		Command:    command,
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

// getPodFrame returns a pod from a DagRun
func (dagRun DAGRun) getPodFrame() core.Pod {
	dag := dagRun.DAG
	return core.Pod{
		TypeMeta: k8sapi.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: k8sapi.ObjectMeta{
			Name:        dagRun.Name,
			Namespace:   dag.Config.Namespace,
			Labels:      dag.Config.Labels,
			Annotations: dag.Config.Annotations,
		},
		Spec: core.PodSpec{
			Volumes:                       nil,
			Containers:                    []core.Container{dagRun.getContainerFrame()},
			EphemeralContainers:           nil,
			RestartPolicy:                 dag.Config.RetryPolicy,
			TerminationGracePeriodSeconds: nil,
			ActiveDeadlineSeconds:         &dag.Config.TimeLimit,
		},
	}
}

// createPod creates and registers a new pod with
func (dagRun *DAGRun) createPod() string {
	dag := dagRun.DAG
	podFrame := dagRun.getPodFrame()
	pod, err := dag.kubeClient.CoreV1().Pods(
		dag.Config.Namespace,
	).Create(
		context.TODO(),
		&podFrame,
		k8sapi.CreateOptions{},
	)
	if err != nil {
		panic(err)
	}
	dagRun.pod = pod
	return pod.Name
}

func (dagRun *DAGRun) monitorPod() watch.Event {
	podsClient := dagRun.DAG.kubeClient.CoreV1().Pods(dagRun.pod.Namespace)
	watcher, err := podsClient.Watch(context.TODO(), k8sapi.ListOptions{})
	if err != nil {
		panic(err)
	}
	result := <-watcher.ResultChan()
	return result
}

// Start starts and monitors the pod and also tracks the logs from the pod
func (dagRun *DAGRun) Start() string {
	podName := dagRun.createPod()
	go dagRun.monitorPod()
	return podName
}

func (dagRun *DAGRun) deletePod() {
	err := dagRun.DAG.kubeClient.CoreV1().Pods(
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
