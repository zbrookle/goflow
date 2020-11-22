package podwatch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"goflow/logs"
	"goflow/podutils"
	"io"
	"strings"

	core "k8s.io/api/core/v1"
	k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// PodWatcher watches events and streams logs from pods while they are running
type PodWatcher struct {
	podName    string
	namespace  string
	kubeClient kubernetes.Interface
	Logs       chan string
	withLogs   bool
	Phase      core.PodPhase
}

func NewPodWatcher(
	name string,
	namespace string,
	client kubernetes.Interface,
	withLogs bool,
) PodWatcher {
	return PodWatcher{name, namespace, client, make(chan string, 1), withLogs, core.PodPending}
}

// podClient returns the api endpoint for pods
func (podWatcher *PodWatcher) podClient() v1.PodInterface {
	return podWatcher.kubeClient.CoreV1().Pods(podWatcher.namespace)
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

func (podWatcher *PodWatcher) watcher() watch.Interface {
	nameSelector := podutils.LabelSelectorString(map[string]string{
		"Name": podWatcher.podName,
	})
	watcher, err := podWatcher.podClient().Watch(
		context.TODO(),
		k8sapi.ListOptions{LabelSelector: nameSelector},
	)
	if err != nil {
		panic(err)
	}
	return watcher
}

// waitForPodState returns when the pod has reached the given state
func (podWatcher *PodWatcher) waitForPodState(watcher watch.Interface, state watch.EventType) {
	for result := range watcher.ResultChan() {
		if result.Type == state {
			break
		}
	}
	logs.InfoLogger.Printf("Pod %s has reached state %s\n", podWatcher.podName, state)
}

// waitForPodAdded returns when the pod has been added
func (podWatcher *PodWatcher) waitForPodAdded(watcher watch.Interface) {
	podWatcher.waitForPodState(watcher, watch.Added)
}

func (podWatcher *PodWatcher) getLogStreamerWithOptions(
	options *core.PodLogOptions,
) (io.ReadCloser, error) {
	req := podWatcher.podClient().GetLogs(podWatcher.podName, options)
	return req.Stream(context.TODO())
}

// getLogsContainerNotFound
func (podWatcher *PodWatcher) getLogsContainerNotFound() (io.ReadCloser, error) {
	pod, err := podWatcher.podClient().Get(context.TODO(), podWatcher.podName, k8sapi.GetOptions{})
	panic("test")
	if err != nil {
		panic(err)
	}
	switch phase := pod.Status.Phase; phase {
	case core.PodSucceeded:
		return podWatcher.getLogStreamerWithOptions(&core.PodLogOptions{Previous: true})
	default:
		return nil, fmt.Errorf("Unexpected error occurred with pod %s", podWatcher.podName)
	}
}

// getLogger returns when logs are ready to be received
func (podWatcher *PodWatcher) getLogger() (io.ReadCloser, error) {
	var logStreamer io.ReadCloser
	for {
		streamer, err := podWatcher.getLogStreamerWithOptions(&core.PodLogOptions{})
		logStreamer = streamer
		if err == nil {
			break
		}
		errorText := err.Error()
		logs.InfoLogger.Println(errorText)
		if strings.Contains(errorText, "not found") {
			panic("test")
			return podWatcher.getLogsContainerNotFound()
		}
	}
	return logStreamer, nil
}

func (podWatcher *PodWatcher) callFuncUntilPodSucceedOrFail(callFunc func()) {
	watcher := podWatcher.watcher()
	for {
		logs.InfoLogger.Println("Calling func...")
		callFunc()
		logs.InfoLogger.Println("Getting events from chan", watcher.ResultChan())
		event, ok := <-watcher.ResultChan()
		logs.InfoLogger.Println(ok)
		if ok {
			phase := eventObjectToPod(event).Status.Phase
			logs.InfoLogger.Printf("Pod switched to phase %s\n", phase)
			if phase == core.PodSucceeded || phase == core.PodFailed {
				podWatcher.Phase = phase
				break
			}
		}
	}
}

func (podWatcher *PodWatcher) readLogsUntilSucceedOrFail(
	logger io.ReadCloser,
	watcher watch.Interface,
) {
	defer logger.Close()
	podWatcher.callFuncUntilPodSucceedOrFail(func() {
		logBuffer := new(bytes.Buffer)
		_, err := io.Copy(logBuffer, logger)
		if err != nil {
			panic(err)
		}
		logString := logBuffer.String()
		if logString != "" && podWatcher.Logs != nil {
			podWatcher.Logs <- logString
		}
	})
}

func (podWatcher *PodWatcher) MonitorPod() {
	watcher := podWatcher.watcher()
	podWatcher.waitForPodAdded(watcher)
	logger, err := podWatcher.getLogger()
	if err != nil {
		panic(err)
	}
	podWatcher.readLogsUntilSucceedOrFail(logger, watcher)
}
