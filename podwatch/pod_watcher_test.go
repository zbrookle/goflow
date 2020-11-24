package podwatch

import (
	"bytes"
	"context"
	"goflow/k8sclient"
	"goflow/logs"
	"goflow/podutils"
	"io"
	"os/exec"
	"strings"
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

func createTestPod(podsClient v1.PodInterface, podName string, namespace string) *core.Pod {
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
	pod, err := podsClient.Create(context.TODO(), &testPodFrame, k8sapi.CreateOptions{})
	if err != nil {
		panic(err)
	}
	return pod
}

func TestEventWatcherAddPod(t *testing.T) {
	defer podutils.CleanUpPods(KUBECLIENT)

	namespace := "default"
	podName := "test-pod-watcher-add-pod"
	podsClient := KUBECLIENT.CoreV1().Pods(namespace)

	watcher := NewPodWatcher(podName, namespace, KUBECLIENT, true)
	defer close(watcher.stopInformerChannel)

	createTestPod(podsClient, podName, namespace)
	pod := <-watcher.informerChans.add
	if pod.Status.Phase != core.PodPending {
		t.Errorf("Pod should be in state running, was in state %s", pod.Status.Phase)
	}
	if pod.Name != podName {
		t.Errorf("Pod should have name %s, but saw name %s", podName, pod.Name)
	}
	if pod.Namespace != namespace {
		t.Errorf("Pod should have namespace %s, but found namespace %s", namespace, pod.Namespace)
	}
}

func TestCallFuncUntilSucceedOrFail(t *testing.T) {
	defer podutils.CleanUpPods(KUBECLIENT)

	namespace := "default"
	podName := "test-pod-succeed-or-fail"
	podsClient := KUBECLIENT.CoreV1().Pods(namespace)
	watcher := NewPodWatcher(podName, namespace, KUBECLIENT, true)

	createTestPod(podsClient, podName, namespace)

	watcher.callFuncUntilPodSucceedOrFail(func() {
		logs.InfoLogger.Println("I'm waiting...")
	})

	if watcher.Phase != core.PodFailed && watcher.Phase != core.PodSucceeded {
		t.Errorf(
			"Expected phase to be %s or %s, not %s",
			core.PodFailed,
			core.PodSucceeded,
			watcher.Phase,
		)
	}
}

func TestGetLogsAfterPodDone(t *testing.T) {
	defer podutils.CleanUpPods(KUBECLIENT)

	namespace := "default"
	podName := "test-pod-get-logs-after-pod-done"
	podsClient := KUBECLIENT.CoreV1().Pods(namespace)
	watcher := NewPodWatcher(podName, namespace, KUBECLIENT, true)

	createdPod := createTestPod(podsClient, podName, namespace)

	watcher.callFuncUntilPodSucceedOrFail(func() {
		logs.InfoLogger.Println("Waiting for pod done...")
	})

	if watcher.Phase != core.PodSucceeded {
		panic("Pod did not succeed")
	}

	logger, err := watcher.getLogger()
	if err != nil {
		panic(err)
	}

	// Read in logs
	logBuffer := new(bytes.Buffer)
	_, err = io.Copy(logBuffer, logger)
	if err != nil {
		panic(err)
	}
	returnedLogString := logBuffer.String()
	cleanLogString := strings.TrimSpace(returnedLogString)
	expectedLogText := createdPod.Spec.Containers[0].Command[1]
	if cleanLogString != expectedLogText {
		t.Errorf("Expected logs would have text %s, but found %s", expectedLogText, cleanLogString)
	}
}
