package dags

import (
	"bytes"
	"context"
	"encoding/json"

	"goflow/logs"
	"goflow/podutils"
	"io"

	core "k8s.io/api/core/v1"
	k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
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
	PodPhase      core.PodPhase
	Logs          chan string
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
	dagRun.PodPhase = pod.Status.Phase
}

// podClient returns the api endpoint for pods
func (dagRun *DAGRun) podClient() v1.PodInterface {
	return dagRun.DAG.kubeClient.CoreV1().Pods(dagRun.DAG.Config.Namespace)
}

// eventObjectToPod returns a pod object from the event result object
func eventObjectToPod(result watch.Event) *core.Pod {
	podObject := &core.Pod{}
	jsonObj, err := json.Marshal(result.Object)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(jsonObj, podObject)
	if err != nil {
		panic(err)
	}
	return podObject
}

func (dagRun *DAGRun) watcher() watch.Interface {
	nameSelector := podutils.LabelSelectorString(map[string]string{
		"Name": dagRun.Name,
	})
	watcher, err := dagRun.podClient().Watch(
		context.TODO(),
		k8sapi.ListOptions{LabelSelector: nameSelector},
	)
	if err != nil {
		panic(err)
	}
	return watcher
}

// waitForPodState returns when the pod has reached the given state
func (dagRun *DAGRun) waitForPodState(watcher watch.Interface, state watch.EventType) {
	logs.InfoLogger.Printf("Wait for pod %s to reach state %s\n", dagRun.Name, state)
	for {
		result := <-watcher.ResultChan()
		if result.Type == state {
			break
		}
	}
}

// waitForPodRunning returns when the pod has been added
func (dagRun *DAGRun) waitForPodRunning(watcher watch.Interface) {
	dagRun.waitForPodState(watcher, watch.Added)
}

// waitForPodDelete returns when the pod has been deleted
func (dagRun *DAGRun) waitForPodDelete(watcher watch.Interface) {
	dagRun.waitForPodState(watcher, watch.Deleted)
}

// getLogger returns when logs are ready to be received
func (dagRun *DAGRun) getLogger() io.ReadCloser {
	var logStreamer io.ReadCloser
	for {
		req := dagRun.podClient().GetLogs(dagRun.pod.Name, &core.PodLogOptions{})
		streamer, err := req.Stream(context.TODO())
		logStreamer = streamer
		if err == nil {
			break
		}
	}
	return logStreamer
}

func (dagRun *DAGRun) readLogsUntilDelete(logger io.ReadCloser, watcher watch.Interface) {
	defer logger.Close()
	for {
		logBuffer := new(bytes.Buffer)
		_, err := io.Copy(logBuffer, logger)
		if err != nil {
			panic(err)
		}
		logString := logBuffer.String()
		if logString != "" && dagRun.Logs != nil {
			dagRun.Logs <- logString
		}
		event, ok := <-watcher.ResultChan()
		if ok {
			phase := eventObjectToPod(event).Status.Phase
			logs.InfoLogger.Printf("Pod switched to phase %s\n", phase)
			if phase == core.PodSucceeded || phase == core.PodFailed {
				dagRun.PodPhase = phase
				break
			}
		}
	}
}

func (dagRun *DAGRun) monitorPod() {
	watcher := dagRun.watcher()
	dagRun.waitForPodRunning(watcher)
	logger := dagRun.getLogger()
	dagRun.readLogsUntilDelete(logger, watcher)
}

// Start starts and monitors the pod and also tracks the logs from the pod
func (dagRun *DAGRun) Start() {
	dagRun.createPod()
	dagRun.monitorPod()
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

func (dagRun *DAGRun) withLogs() {
	dagRun.Logs = make(chan string, 1)
}
