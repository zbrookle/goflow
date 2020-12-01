package run

import (
	"context"
	"fmt"

	"goflow/logs"

	dagconfig "goflow/dag/config"
	"goflow/k8s/pod/event/holder"
	podwatch "goflow/k8s/pod/watch"

	"time"

	"strings"

	core "k8s.io/api/core/v1"
	k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// DAGRun is a single run of a given dag - corresponds with a kubernetes pod
type DAGRun struct {
	Name          string
	Config        *dagconfig.DAGConfig
	ExecutionDate k8sapi.Time // This is the date that will be passed to the pod that runs
	StartTime     k8sapi.Time
	EndTime       k8sapi.Time
	pod           *core.Pod
	withLogs      bool
	kubeClient    kubernetes.Interface
	watcher       *podwatch.PodWatcher
	holder        *holder.ChannelHolder
}

func cleanK8sName(name string) string {
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ReplaceAll(name, ":", "-")
	name = strings.ReplaceAll(name, " ", "")
	name = strings.ReplaceAll(name, "+", "plus")
	name = strings.ToLower(name)
	return name
}

// NewDAGRun returns a new instance of DAGRun
func NewDAGRun(
	executionDate time.Time,
	dagConfig *dagconfig.DAGConfig,
	withLogs bool,
	kubeClient kubernetes.Interface,
	channelHolder *holder.ChannelHolder,
) *DAGRun {
	podName := cleanK8sName(dagConfig.Name + executionDate.String())
	return &DAGRun{
		Name:   podName,
		Config: dagConfig,
		ExecutionDate: k8sapi.Time{
			Time: executionDate,
		},
		StartTime: k8sapi.Time{
			Time: time.Now(),
		},
		EndTime: k8sapi.Time{
			Time: time.Time{},
		},
		withLogs:   withLogs,
		kubeClient: kubeClient,
		watcher: podwatch.NewPodWatcher(
			podName,
			dagConfig.Namespace,
			kubeClient,
			withLogs,
			channelHolder,
		),
		holder: channelHolder,
	}
}

func (dagRun *DAGRun) getContainerFrame() core.Container {
	return core.Container{
		Name:            "task",
		Image:           dagRun.Config.DockerImage,
		Command:         dagRun.Config.Command,
		Args:            nil,
		WorkingDir:      "",
		EnvFrom:         nil,
		Env:             nil,
		VolumeMounts:    nil,
		VolumeDevices:   nil,
		ImagePullPolicy: "IfNotPresent",
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
	labels := copyStringMap(dagRun.Config.Labels)
	labels["Name"] = dagRun.Name
	labels["App"] = "goflow"
	return core.Pod{
		TypeMeta: k8sapi.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: k8sapi.ObjectMeta{
			Name:        dagRun.Name,
			Namespace:   dagRun.Config.Namespace,
			Labels:      labels,
			Annotations: dagRun.Config.Annotations,
		},
		Spec: core.PodSpec{
			Volumes:                       nil,
			Containers:                    []core.Container{dagRun.getContainerFrame()},
			EphemeralContainers:           nil,
			RestartPolicy:                 dagRun.Config.RetryPolicy,
			TerminationGracePeriodSeconds: nil,
			ActiveDeadlineSeconds:         &dagRun.Config.TimeLimit,
		},
	}
}

// createPod creates and registers a new pod with
func (dagRun *DAGRun) createPod() {
	podFrame := dagRun.getPodFrame()
	logs.InfoLogger.Printf("Creating pod %s...\n", podFrame.Name)
	pod, err := dagRun.podClient().Create(
		context.TODO(),
		&podFrame,
		k8sapi.CreateOptions{},
	)
	if err != nil {
		panic(err)
	}
	logs.InfoLogger.Printf(
		"Pod '%s' created in namespace '%s'\n",
		podFrame.Name,
		podFrame.Namespace,
	)
	dagRun.pod = pod
}

// podClient returns the api endpoint for pods
func (dagRun *DAGRun) podClient() v1.PodInterface {
	return dagRun.kubeClient.CoreV1().Pods(dagRun.Config.Namespace)
}

// Run runs the pods and monitoring methods
func (dagRun *DAGRun) Run() {
	podFrame := dagRun.getPodFrame()
	dagRun.holder.AddChannelGroup(podFrame.Name)
	go dagRun.watcher.MonitorPod() // Start monitoring before the pod is actually running
	dagRun.createPod()
}

// Start runs the dagrun and waits for the monitoring to finish
func (dagRun *DAGRun) Start() {
	go dagRun.Run()
	dagRun.watcher.WaitForMonitorDone()
}

// Logs returns the channel holding the watcher's logs
func (dagRun *DAGRun) Logs() *chan string {
	return &dagRun.watcher.Logs
}

// DeletePod deletes the dag run's associated pod
func (dagRun *DAGRun) DeletePod() {
	logs.InfoLogger.Printf(
		"Deleting pod %s, in namespace %s",
		dagRun.pod.Name,
		dagRun.pod.Namespace,
	)
	err := dagRun.podClient().Delete(
		context.TODO(),
		dagRun.Name,
		k8sapi.DeleteOptions{},
	)
	if err != nil {
		panic(err)
	}
}

// MostRecentPod returns the pod run for this dag run
func (dagRun *DAGRun) MostRecentPod() (core.Pod, error) {
	if dagRun.pod == nil {
		return core.Pod{}, fmt.Errorf("pod %s has not been created yet", dagRun.Name)
	}
	return *dagRun.pod, nil
}

// TRY COUNTING EVENT STATES -- USE this as rate limiting - if pod is pending for too long
