package watch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"goflow/k8s/pod/event/holder"
	"goflow/logs"
	"io"
	"strings"

	core "k8s.io/api/core/v1"
	k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// Add func Channels cache

// PodWatcher watches events and streams logs from pods while they are running
type PodWatcher struct {
	podName        string
	namespace      string
	kubeClient     kubernetes.Interface
	Logs           chan string
	withLogs       bool
	Phase          core.PodPhase
	informerChans  *holder.ChannelHolder
	monitoringDone chan struct{}
}

func getPodFromInterface(obj interface{}) *core.Pod {
	pod, ok := obj.(*core.Pod)
	if !ok {
		panic(fmt.Sprintf("Expected %T, but go %T", &core.Pod{}, obj))
	}
	return pod
}

func podReadyToLog(pod *core.Pod) bool {
	return (pod.Status.Phase == core.PodRunning) || (pod.Status.Phase == core.PodSucceeded) ||
		(pod.Status.Phase == core.PodFailed)
}

// NewPodWatcher returns a new pod watcher
func NewPodWatcher(
	name string,
	namespace string,
	client kubernetes.Interface,
	withLogs bool,
	channelGroupHolder *holder.ChannelHolder,
) *PodWatcher {
	return &PodWatcher{
		podName:       name,
		namespace:     namespace,
		kubeClient:    client,
		Logs:          make(chan string, 1),
		withLogs:      withLogs,
		informerChans: channelGroupHolder,
	}
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

// waitForPodAdded returns when the pod has been added
func (podWatcher *PodWatcher) waitForPodAdded() {
	pod := <-podWatcher.informerChans.GetChannelGroup(podWatcher.podName).Ready
	podWatcher.Phase = pod.Status.Phase
}

func (podWatcher *PodWatcher) getLogStreamerWithOptions(
	options *core.PodLogOptions,
) (io.ReadCloser, error) {
	req := podWatcher.podClient().GetLogs(podWatcher.podName, options)
	return req.Stream(context.TODO())
}

// getLogsContainerNotFound
func (podWatcher *PodWatcher) getLogsContainerNotFound() (io.ReadCloser, error) {
	return podWatcher.getLogStreamerWithOptions(&core.PodLogOptions{Previous: true})
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
		if strings.Contains(errorText, "not found") {
			return podWatcher.getLogsContainerNotFound()
		}
	}
	return logStreamer, nil
}

func (podWatcher *PodWatcher) getPodFromK8s() *core.Pod {
	pod, err := podWatcher.podClient().Get(context.TODO(), podWatcher.podName, k8sapi.GetOptions{})
	if err != nil {
		panic(err)
	}
	return pod
}

func isPodComplete(pod *core.Pod) bool {
	return pod.Status.Phase == core.PodFailed || pod.Status.Phase == core.PodSucceeded
}

func (podWatcher *PodWatcher) callFuncUntilPodSucceedOrFail(callFunc func()) {
	if podWatcher.Phase == core.PodFailed || podWatcher.Phase == core.PodSucceeded {
		callFunc()
		return
	}
	for {
		callFunc()
		logs.InfoLogger.Println("Waiting for ready channel...")
		pod, ok := <-podWatcher.informerChans.GetChannelGroup(podWatcher.podName).Update
		if ok {
			phase := pod.Status.Phase
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

func (podWatcher *PodWatcher) setMonitorDone() {
	podWatcher.monitoringDone <- struct{}{}
}

// MonitorPod collects pod logs until the pod terminates
func (podWatcher *PodWatcher) MonitorPod() {
	defer podWatcher.setMonitorDone()

	podWatcher.waitForPodAdded()
	logger, err := podWatcher.getLogger()
	if err != nil {
		panic(err)
	}
	podWatcher.readLogsUntilSucceedOrFail(logger)
}

// WaitForMonitorDone returns when the watcher is done monitoring
func (podWatcher *PodWatcher) WaitForMonitorDone() {
	<-podWatcher.monitoringDone
}
