package inform

import (
	"goflow/internal/k8s/pod/event/holder"
	podutils "goflow/internal/k8s/pod/utils"
	"goflow/internal/testutils"

	"testing"

	core "k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes/fake"
)

const testPodName = "test"

func TestPodReadyForLogging(t *testing.T) {
	table := []struct {
		phase core.PodPhase
	}{
		{core.PodRunning},
		{core.PodFailed},
		{core.PodSucceeded},
	}
	for _, phaseStruct := range table {
		func() {
			KUBECLIENT := fake.NewSimpleClientset()
			channelHolder := holder.New()
			taskInformer := New(KUBECLIENT, channelHolder)
			podName := "test-pod-watcher-add-pod"
			namespace := "default"

			go taskInformer.Start()
			defer taskInformer.Stop()

			channelHolder.AddChannelGroup(podName)
			podsClient, watcher := testutils.GetPodClientWithTestWatcher(KUBECLIENT, namespace)

			createdPod := podutils.CreateTestPod(podsClient, podName, namespace, core.PodPending)
			watcher.Add(createdPod)
			createdPod.Status.Phase = phaseStruct.phase

			t.Log("Waiting for pod ready...")
			pod := <-channelHolder.GetChannelGroup(podName).Ready
			t.Log("Pod value taken from channel...")
			if !podReadyToLog(pod) {
				t.Errorf("Pod is not ready to log!")
			}
			if pod.Name != podName {
				t.Errorf("Pod should have name %s, but saw name %s", podName, pod.Name)
			}
			if pod.Namespace != namespace {
				t.Errorf(
					"Pod should have namespace %s, but found namespace %s",
					namespace,
					pod.Namespace,
				)
			}
		}()

	}

}
