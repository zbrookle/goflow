package watch

import (
	"bytes"
	"context"
	"encoding/json"

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

// NewPodWatcher returns a new pod watcher
func NewPodWatcher(
	name string,
	namespace string,
	client kubernetes.Interface,
	withLogs bool,
	channelGroupHolder *holder.ChannelHolder,
) *PodWatcher {
	return &PodWatcher{
		podName:        name,
		namespace:      namespace,
		kubeClient:     client,
		Logs:           make(chan string, 1),
		withLogs:       withLogs,
		informerChans:  channelGroupHolder,
		monitoringDone: make(chan struct{}, 1),
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
	logs.InfoLogger.Printf("Waiting for pod %s to be added...\n", podWatcher.podName)
	if !podWatcher.informerChans.Contains(podWatcher.podName) {
		logs.ErrorLogger.Printf("Channels not found for pod %s\n", podWatcher.podName)
	}
	pod := <-podWatcher.informerChans.GetChannelGroup(podWatcher.podName).Ready
	podWatcher.Phase = pod.Status.Phase
	logs.InfoLogger.Printf("Pod %s added\n", podWatcher.podName)
}

func (podWatcher *PodWatcher) getLogStreamerWithOptions(
	options *core.PodLogOptions,
) (io.ReadCloser, error) {
	req := podWatcher.podClient().GetLogs(podWatcher.podName, options)
	return req.Stream(context.Background())
}

// getLogsContainerNotFound
func (podWatcher *PodWatcher) getLogsContainerNotFound() (io.ReadCloser, error) {
	return podWatcher.getLogStreamerWithOptions(&core.PodLogOptions{Previous: true})
}

// getLogger returns when logs are ready to be received
func (podWatcher *PodWatcher) getLogger() (io.ReadCloser, error) {
	logs.InfoLogger.Printf("Retrieving logger for pod %s...\n", podWatcher.podName)
	var logStreamer io.ReadCloser
	for {
		streamer, err := podWatcher.getLogStreamerWithOptions(&core.PodLogOptions{})
		logStreamer = streamer
		if err == nil {
			break
		}
		errorText := err.Error()
		if strings.Contains(errorText, "not found") {
			logs.InfoLogger.Printf(
				"Container not found for pod %s, handling...\n",
				podWatcher.podName,
			)
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
		logs.InfoLogger.Println("Waiting for pod update...")
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

func getStringFromLogger(
	logger io.ReadCloser,
	logChan chan string,
	podName string,
) (addedLogs bool) {
	addedLogs = false
	logBuffer := new(bytes.Buffer)
	_, err := io.Copy(logBuffer, logger)
	if err != nil {
		panic(err)
	}
	logString := logBuffer.String()
	logs.InfoLogger.Println("Found log", logString, "for pod", podName)
	logs.InfoLogger.Println("Logs channel:", logChan, "for pod", podName)
	if logString != "" {
		logs.InfoLogger.Println("Added log:", logString, "for pod", podName)
		addedLogs = true
		logChan <- logString
	}
	return
}

func (podWatcher *PodWatcher) readLogsUntilSucceedOrFail(
	logger io.ReadCloser,
) {
	logs.InfoLogger.Println(podWatcher.podName)
	defer logger.Close()
	addedLogs := false
	podWatcher.callFuncUntilPodSucceedOrFail(func() {
		if getStringFromLogger(logger, podWatcher.Logs, podWatcher.podName) {
			addedLogs = true
		}
	})
	if !addedLogs {
		addedLogs = getStringFromLogger(logger, podWatcher.Logs, podWatcher.podName)
		if !addedLogs {
			logs.InfoLogger.Printf("No logs retrieved for pod %s\n", podWatcher.podName)
		}
	}
}

func (podWatcher *PodWatcher) setMonitorDone() {
	logs.InfoLogger.Printf("Monitoring for pod %s done", podWatcher.podName)
	podWatcher.monitoringDone <- struct{}{}
}

// MonitorPod collects pod logs until the pod terminates
func (podWatcher *PodWatcher) MonitorPod() {
	defer podWatcher.setMonitorDone()
	logs.InfoLogger.Printf("Beginning to monitor pod %s\n", podWatcher.podName)
	podWatcher.waitForPodAdded()
	logger, err := podWatcher.getLogger()
	if err != nil {
		panic(err)
	}
	podWatcher.readLogsUntilSucceedOrFail(logger)
}

// WaitForMonitorDone returns when the watcher is done monitoring
func (podWatcher *PodWatcher) WaitForMonitorDone() {
	logs.InfoLogger.Printf("Waiting for pod %s to be done\n", podWatcher.podName)
	<-podWatcher.monitoringDone
}
