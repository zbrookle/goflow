package watch

import (
	"bytes"
	"goflow/k8s/pod/event/holder"
	podutils "goflow/k8s/pod/utils"
	"goflow/logs"
	"goflow/testutils"
	"io"
	"os/exec"
	"strings"
	"testing"

	core "k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes/fake"
)

func displayPods() {
	exec.Command("kubectl", "get", "pods")
}

func TestCallFuncUntilSucceedOrFail(t *testing.T) {
	table := []struct {
		finalPhase core.PodPhase
	}{
		{core.PodFailed},
		{core.PodSucceeded},
	}
	for _, testCase := range table {
		func() {
			client := fake.NewSimpleClientset()

			namespace := "default"
			podName := "test-pod-succeed-or-fail"
			holder := holder.New()
			podWatcher := NewPodWatcher(podName, namespace, client, true, &holder)
			podsClient, _ := testutils.GetPodClientWithTestWatcher(client, namespace)
			testPod := podutils.CreateTestPod(podsClient, podName, namespace, "")
			t.Log("Test pod created")

			updatedPod := testPod.DeepCopy()
			updatedPod.Status.Phase = testCase.finalPhase
			holder.AddChannelGroup(podName)
			holder.GetChannelGroup(podName).Update <- updatedPod

			podWatcher.callFuncUntilPodSucceedOrFail(func() {
				logs.InfoLogger.Println("I'm waiting...")
			})

			if podWatcher.Phase != testCase.finalPhase {
				t.Errorf(
					"Expected phase to be %s, not %s",
					testCase.finalPhase,
					podWatcher.Phase,
				)
			}
		}()
	}

}

func TestGetLogsAfterPodDone(t *testing.T) {
	table := []struct {
		finalPhase core.PodPhase
	}{
		{core.PodSucceeded},
		{core.PodFailed},
	}
	for _, testCase := range table {
		func() {
			client := fake.NewSimpleClientset()

			namespace := "default"
			podName := "test-pod-get-logs-after-pod-done"
			podsClient, _ := testutils.GetPodClientWithTestWatcher(client, namespace)
			channelHolder := holder.New()
			watcher := NewPodWatcher(podName, namespace, client, true, &channelHolder)
			watcher.informerChans.AddChannelGroup(podName)

			createdPod := podutils.CreateTestPod(podsClient, podName, namespace, "")
			podCopy := createdPod.DeepCopy()
			podCopy.Status.Phase = testCase.finalPhase
			watcher.informerChans.GetChannelGroup(podName).Update <- podCopy

			watcher.callFuncUntilPodSucceedOrFail(func() {
				logs.InfoLogger.Println("Waiting for pod done...")
			})

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
			if cleanLogString != expectedLogText && cleanLogString != "fake logs" {
				t.Errorf(
					"Expected logs would have text %s, but found %s",
					expectedLogText,
					cleanLogString,
				)
			}
		}()
	}

}
