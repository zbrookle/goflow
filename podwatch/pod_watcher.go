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

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
)

type funcChannels struct {
	add    chan *core.Pod
	update chan *core.Pod
	remove chan *core.Pod
}

// PodWatcher watches events and streams logs from pods while they are running
type PodWatcher struct {
	podName         string
	namespace       string
	kubeClient      kubernetes.Interface
	Logs            chan string
	withLogs        bool
	Phase           core.PodPhase
	informer        cache.SharedInformer
	informerChans   funcChannels
	informerStopper chan struct{}
}

func newWatcher(podName string, podClient v1.PodInterface) watch.Interface {
	nameSelector := podutils.LabelSelectorString(map[string]string{
		"Name": podName,
	})
	watcher, err := podClient.Watch(
		context.TODO(),
		k8sapi.ListOptions{LabelSelector: nameSelector},
	)
	if err != nil {
		panic(err)
	}
	return watcher
}

func getPodFromInterface(obj interface{}) *core.Pod {
	pod, ok := obj.(*core.Pod)
	if !ok {
		panic(fmt.Sprintf("Expected %T, but go %T", &core.Pod{}, obj))
	}
	return pod
}

func getSharedInformer(
	client kubernetes.Interface,
	name string,
	namespace string,
	addFuncChannel chan int,
) (cache.SharedInformer, funcChannels) {
	listWatcher := cache.NewListWatchFromClient(
		client.CoreV1().RESTClient(),
		"pods",
		namespace,
		// fields.Everything(),
		fields.OneTermEqualSelector("metadata.name", name),
	)
	informer := cache.NewSharedInformer(listWatcher, &core.Pod{}, 0)
	// factory := informers.NewSharedInformerFactoryWithOptions(
	// 	client,
	// 	0,
	// 	informers.WithNamespace(namespace),
	// )
	// informer := factory.Core().V1().Pods().Informer()

	channels := funcChannels{
		make(chan *core.Pod, 1),
		make(chan *core.Pod, 1),
		make(chan *core.Pod),
	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			logs.InfoLogger.Printf("Call add pod func")
			pod := getPodFromInterface(obj)
			logs.InfoLogger.Printf("Pod with name %s, in phase %s", pod.Name, pod.Status.Phase)
			// for _, status := range pod.Status.ContainerStatuses {
			// 	logs.InfoLogger.Printf("Container in phase %s", status.State.String())
			// }
			channels.add <- getPodFromInterface(obj)
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			logs.InfoLogger.Println("Call pod update function")
			channels.update <- getPodFromInterface(new)
		},
		DeleteFunc: func(obj interface{}) {
			channels.update <- getPodFromInterface(obj)
		},
	})
	return informer, channels
}

// NewPodWatcher returns a new pod watcher
func NewPodWatcher(
	name string,
	namespace string,
	client kubernetes.Interface,
	withLogs bool,
) *PodWatcher {
	addFuncChannel := make(chan int, 1)
	stopChannel := make(chan struct{})
	informer, channels := getSharedInformer(client, name, namespace, addFuncChannel)
	go informer.Run(stopChannel)
	return &PodWatcher{
		name,
		namespace,
		client,
		make(chan string, 1),
		withLogs,
		core.PodPending,
		informer,
		channels,
		stopChannel,
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
	logs.InfoLogger.Println("Waiting for pod!")
	<-podWatcher.informerChans.add
	logs.InfoLogger.Println("Pod added!!!!")
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
	for {
		callFunc()
		// myChan := podWatcher.eventChan
		// logs.InfoLogger.Println("Getting events from chan", myChan)
		// if podWatcher.eventChan == nil {
		// 	panic("Channel is nil!!!")
		// }
		logs.InfoLogger.Println("Waiting for update channel...")
		pod, ok := <-podWatcher.informerChans.update
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

// MonitorPod collects pod logs until the pod terminates
func (podWatcher *PodWatcher) MonitorPod() {
	podWatcher.waitForPodAdded()
	logger, err := podWatcher.getLogger()
	if err != nil {
		panic(err)
	}
	podWatcher.readLogsUntilSucceedOrFail(logger)
}
