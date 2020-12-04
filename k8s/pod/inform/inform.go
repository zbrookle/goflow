package inform

import (
	"goflow/k8s/pod/event/channel"
	"goflow/k8s/pod/event/holder"
	"goflow/logs"
	"time"

	"fmt"

	core "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// TaskInformer is a custom informer that updates the pod states while in the channel holder
type TaskInformer struct {
	podInformer         cache.SharedInformer
	channelHolder       *holder.ChannelHolder
	stopInformerChannel chan struct{}
}

// New returns a new informer to be used in updating the channels in the holder
func New(
	client kubernetes.Interface,
	channelHolder *holder.ChannelHolder,
) TaskInformer {
	factory := informers.NewSharedInformerFactory(client, 2*time.Second)
	sharedInformer := factory.Core().V1().Pods().Informer()
	taskInformer := TaskInformer{sharedInformer, channelHolder, make(chan struct{})}

	sharedInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := getPodFromInterface(obj)
			if taskInformer.channelHolder.Contains(pod.Name) && podReadyToLog(pod) {
				select {
				case taskInformer.getChannelGroup(pod.Name).Ready <- pod:
					logs.InfoLogger.Printf(
						"Pod with name %s added and ready in phase %s\n",
						pod.Name,
						pod.Status.Phase,
					)
				default:
					logs.InfoLogger.Printf(
						"Pod with name %s already added from update",
						pod.Name,
					)
				}

			}
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			oldPod := getPodFromInterface(old)
			newPod := getPodFromInterface(new)
			if !taskInformer.channelHolder.Contains(newPod.Name) {
				return
			}
			if podReadyToLog(newPod) {
				select {
				case taskInformer.getChannelGroup(newPod.Name).Ready <- newPod:
					logs.InfoLogger.Printf(
						"Pod %s updated to ready in phase %s",
						newPod.Name,
						newPod.Status.Phase,
					)
				default:
					logs.InfoLogger.Printf(
						"Pod with name %s already added from update\n",
						newPod.Name,
					)
				}

			}
			if oldPod.Status.Phase != newPod.Status.Phase {
				taskInformer.getChannelGroup(newPod.Name).Update <- newPod
				logs.InfoLogger.Printf(
					"Pod %s updated from phase %s to phase %s",
					newPod.Name,
					oldPod.Status.Phase,
					newPod.Status.Phase,
				)
			}
		},
		DeleteFunc: func(obj interface{}) {
			pod := getPodFromInterface(obj)
			logs.InfoLogger.Printf("Pod %s was successfully deleted\n", pod.Name)
		},
	})
	return taskInformer
}

func podReadyToLog(pod *core.Pod) bool {
	return (pod.Status.Phase == core.PodRunning) || (pod.Status.Phase == core.PodSucceeded) ||
		(pod.Status.Phase == core.PodFailed)
}

func getPodFromInterface(obj interface{}) *core.Pod {
	pod, ok := obj.(*core.Pod)
	if !ok {
		panic(fmt.Sprintf("Expected %T, but go %T", &core.Pod{}, obj))
	}
	return pod
}

func (taskInformer *TaskInformer) getChannelGroup(podName string) *channel.FuncChannelGroup {
	return taskInformer.channelHolder.GetChannelGroup(podName)
}

// Stop stops the running informer
func (taskInformer *TaskInformer) Stop() {
	taskInformer.stopInformerChannel <- struct{}{}
}

// Start starts the informer
func (taskInformer *TaskInformer) Start() {
	go taskInformer.podInformer.Run(taskInformer.stopInformerChannel)
}
