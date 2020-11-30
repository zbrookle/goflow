package inform

import (
	// "context"
	"goflow/k8s/pod/event/holder"
	podutils "goflow/k8s/pod/utils"

	"testing"

	core "k8s.io/api/core/v1"
	// k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	fakecore "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	testcore "k8s.io/client-go/testing"
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

			podsClient := KUBECLIENT.CoreV1().Pods(namespace)
			fakePodClient := podsClient.(*fakecore.FakePods)
			watcher := watch.NewFake()
			fakePodClient.Fake.PrependWatchReactor(
				"pods",
				testcore.DefaultWatchReactor(watcher, nil),
			)

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
