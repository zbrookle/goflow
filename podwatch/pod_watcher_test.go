package podwatch

import (
	"context"
	"goflow/k8sclient"
	"goflow/logs"
	"goflow/podutils"
	"os/exec"
	"testing"

	core "k8s.io/api/core/v1"
	k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var KUBECLIENT kubernetes.Interface

func TestMain(m *testing.M) {
	KUBECLIENT = k8sclient.CreateKubeClient()
	podutils.CleanUpPods(KUBECLIENT)
	m.Run()
}

func displayPods() {
	exec.Command("kubectl", "get", "pods")
}

func createTestPod(podsClient v1.PodInterface, podName string, namespace string) {
	testPodFrame := core.Pod{
		TypeMeta: k8sapi.TypeMeta{
			Kind:       "Pod",
			APIVersion: "V1",
		},
		ObjectMeta: k8sapi.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Labels:    map[string]string{podutils.AppSelectorKey: podutils.AppName},
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:            "task",
					Image:           "busybox",
					Command:         []string{"echo", "123 test"},
					ImagePullPolicy: core.PullIfNotPresent,
				},
			},
			RestartPolicy: core.RestartPolicyNever,
		},
	}
	_, err := podsClient.Create(context.TODO(), &testPodFrame, k8sapi.CreateOptions{})
	if err != nil {
		panic(err)
	}
}

func TestEventWatcherAddPod(t *testing.T) {
	defer podutils.CleanUpPods(KUBECLIENT)

	namespace := "default"
	podName := "test-pod"
	podsClient := KUBECLIENT.CoreV1().Pods(namespace)

	watcher := NewPodWatcher(podName, namespace, KUBECLIENT, true)
	defer close(watcher.informerStopper)

	createTestPod(podsClient, podName, namespace)
	pod := <-watcher.informerChans.add
	if pod.Status.Phase != core.PodRunning {
		t.Error("Pod should be in state running")
	}
}

func TestFindContainerCompleteEvent(t *testing.T) {
	defer podutils.CleanUpPods(KUBECLIENT)

	namespace := "default"
	podName := "test-pod"
	podsClient := KUBECLIENT.CoreV1().Pods(namespace)
	watcher := NewPodWatcher(podName, namespace, KUBECLIENT, true)
	defer close(watcher.informerStopper)

	createTestPod(podsClient, podName, namespace)

	watcher.callFuncUntilPodSucceedOrFail(func() {
		logs.InfoLogger.Println("I'm waiting...")
	})
}
