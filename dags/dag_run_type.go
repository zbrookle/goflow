package dags

import (
	"context"

	"goflow/logs"

	"goflow/podwatch"

	core "k8s.io/api/core/v1"
	k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// DAGRun is a single run of a given dag - corresponds with a kubernetes pod
type DAGRun struct {
	Name          string
	DAG           *DAG
	ExecutionDate k8sapi.Time // This is the date that will be passed to the pod that runs
	StartTime     k8sapi.Time
	EndTime       k8sapi.Time
	pod           *core.Pod
	withLogs      bool
	watcher       podwatch.PodWatcher
}

func (dagRun *DAGRun) getContainerFrame() core.Container {
	return core.Container{
		Name:       "task",
		Image:      dagRun.DAG.Config.DockerImage,
		Command:    dagRun.DAG.Config.Command,
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

func copyStringMap(mapToCopy map[string]string) map[string]string {
	copy := make(map[string]string)
	for key := range mapToCopy {
		copy[key] = mapToCopy[key]
	}
	return copy
}

// getPodFrame returns a pod from a DagRun
func (dagRun *DAGRun) getPodFrame() core.Pod {
	dag := dagRun.DAG
	labels := copyStringMap(dag.Config.Labels)
	labels["Name"] = dagRun.Name
	labels["App"] = "goflow"
	return core.Pod{
		TypeMeta: k8sapi.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: k8sapi.ObjectMeta{
			Name:        dagRun.Name,
			Namespace:   dag.Config.Namespace,
			Labels:      labels,
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
func (dagRun *DAGRun) createPod() {
	logs.InfoLogger.Println("Creating pod...")
	podFrame := dagRun.getPodFrame()
	pod, err := dagRun.podClient().Create(
		context.TODO(),
		&podFrame,
		k8sapi.CreateOptions{},
	)
	if err != nil {
		panic(err)
	}
	dagRun.pod = pod
}

// podClient returns the api endpoint for pods
func (dagRun *DAGRun) podClient() v1.PodInterface {
	return dagRun.DAG.kubeClient.CoreV1().Pods(dagRun.DAG.Config.Namespace)
}

// Start starts and monitors the pod and also tracks the logs from the pod
func (dagRun *DAGRun) Start() {
	dagRun.createPod()
	dagRun.watcher = podwatch.NewPodWatcher(
		dagRun.pod.Name,
		dagRun.pod.Namespace,
		dagRun.DAG.kubeClient,
		dagRun.withLogs,
	)
	dagRun.watcher.MonitorPod()
}

func (dagRun *DAGRun) Logs() *chan string {
	return &dagRun.watcher.Logs
}

func (dagRun *DAGRun) deletePod() {
	err := dagRun.podClient().Delete(
		context.TODO(),
		dagRun.Name,
		k8sapi.DeleteOptions{},
	)
	if err != nil {
		panic(err)
	}
}

// func (dagRun *DAGRun) withLogs() {
// 	dagRun.Logs = make(chan string, 1)
// }

// TRY COUNTING EVENT STATES -- USE this as rate limiting - if pod is pending for too long
