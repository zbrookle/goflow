package dags

import (
	"context"
	"goflow/jsonpanic"
	"goflow/k8sclient"
	"goflow/podutils"
	"strings"
	"testing"

	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreatePod(t *testing.T) {
	defer podutils.CleanUpPods(KUBECLIENT)
	dagRun := createDagRun(getTestDate(), getTestDAGFakeClient(), false)
	dagRun.createPod()
	foundPod, err := dagRun.DAG.kubeClient.CoreV1().Pods(
		dagRun.DAG.Config.Namespace,
	).Get(
		context.TODO(),
		dagRun.pod.Name,
		v1.GetOptions{},
	)
	if err != nil {
		panic(err)
	}
	foundPodValue := jsonpanic.JSONPanic(*foundPod)
	expectedValue := jsonpanic.JSONPanic(*dagRun.pod)
	if foundPodValue != expectedValue {
		t.Error("Expected:", expectedValue)
		t.Error("Found:", foundPodValue)
	}
}

func TestStartPod(t *testing.T) {
	// Test with logs and without logs
	realClient := k8sclient.CreateKubeClient()
	tables := []struct {
		name     string
		withLogs bool
	}{
		{"Without Logs", false},
		{"With Logs", true},
	}
	for _, table := range tables {
		t.Logf("Test case: %s", table.name)
		func() {
			defer podutils.CleanUpPods(realClient)
			dagRun := createDagRun(getTestDate(), getTestDAGRealClient(), table.withLogs)
			dagRun.Start()

			// Test for dag completion in state of dag
			if (dagRun.watcher.Phase != core.PodSucceeded) &&
				(dagRun.watcher.Phase != core.PodFailed) {
				t.Errorf(
					"A finished dagRun should be in phase %s or state %s, but found in state %s",
					core.PodSucceeded,
					core.PodFailed,
					dagRun.watcher.Phase,
				)
			}

			// Test for log output if logs enabled
			if table.withLogs {
				logMsg := <-*dagRun.Logs()
				expectedLogMessage := dagRun.DAG.Config.Command[1]
				logMsg = strings.ReplaceAll(logMsg, "\n", "")
				if logMsg != expectedLogMessage {
					t.Errorf(
						"Expected log message %s, found log message %s",
						expectedLogMessage,
						logMsg,
					)
				}
			}
		}()

	}

}

func TestDeletePod(t *testing.T) {
	defer podutils.CleanUpPods(KUBECLIENT)
	dagRun := createDagRun(getTestDate(), getTestDAGFakeClient(), false)
	podFrame := dagRun.getPodFrame()
	podsClient := KUBECLIENT.CoreV1().Pods(dagRun.DAG.Config.Namespace)

	createdPod, err := podsClient.Create(context.TODO(), &podFrame, v1.CreateOptions{})
	dagRun.pod = createdPod
	if err != nil {
		panic(err)
	}
	dagRun.deletePod()
	list, err := podsClient.List(context.TODO(), v1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, pod := range list.Items {
		if jsonpanic.JSONPanic(*createdPod) == jsonpanic.JSONPanic(pod) {
			t.Errorf("Pod %s should have been deleted", createdPod.Name)
		}
	}
}

// func TestFindContainerCompleteEvent(t *testing.T) {
// 	logs.InfoLogger.Println("Testing Container complete event")
// 	defer podutils.CleanUpPods(KUBECLIENT)
// 	dagRun := createDagRun(getTestDate(), getTestDAGFakeClient())
// 	dagRun.createPod()
// 	dagRun.callFuncUntilPodSucceedOrFail(func() {
// 		logs.InfoLogger.Println("I'm waiting...")
// 	})
// 	// dagRun.getLogsContainerNotFound()
// 	// panic("test")
// }
