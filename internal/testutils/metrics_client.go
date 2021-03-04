package testutils

import (
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/testing"
)

// RegisterContainerStatusesToPods continually adds fake container statuses onto all pod containers
func RegisterContainerStatusesToPods(
	kubeClient *fake.Clientset,
) {
	podReactorFunc := func(action testing.Action) (handled bool, ret runtime.Object, err error) {
		createAction := action.(testing.CreateActionImpl)
		objMeta, err := meta.Accessor(createAction.GetObject())
		if err != nil {
			panic(err)
		}
		pod := objMeta.(*core.Pod)
		started := true
		pod.Status.ContainerStatuses = append(
			pod.Status.ContainerStatuses,
			core.ContainerStatus{Started: &started, Ready: true, Name: "task"},
		)
		kubeClient.Tracker().Add(pod)
		return true, pod, nil
	}

	kubeClient.PrependReactor("create", "pods", podReactorFunc)
}
