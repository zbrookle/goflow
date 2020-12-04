package run

import (
	"context"
	dagconfig "goflow/dag/config"
	"goflow/jsonpanic"

	"goflow/k8s/pod/event/holder"
	podutils "goflow/k8s/pod/utils"
	"strings"
	"testing"

	"time"

	core "k8s.io/api/core/v1"
	k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes/fake"
)

func getTestDate() time.Time {
	return time.Date(2019, 1, 1, 0, 0, 0, 0, time.Now().Location())
}

func getTestDAGConfig(name string, command []string) *dagconfig.DAGConfig {
	if len(command) == 0 {
		command = []string{"echo", "\"Hello world!!!!!!!\""}
	}
	return &dagconfig.DAGConfig{
		Name:          name,
		Namespace:     "default",
		Schedule:      "* * * * *",
		DockerImage:   "busybox",
		RetryPolicy:   "Never",
		Command:       command,
		TimeLimit:     20,
		MaxActiveRuns: 1,
		StartDateTime: "2019-01-01",
		EndDateTime:   "",
	}
}

func TestCreatePod(t *testing.T) {
	client := fake.NewSimpleClientset()
	defer podutils.CleanUpEnvironment(client)
	dagRun := NewDAGRun(
		getTestDate(),
		getTestDAGConfig("test-create-pod", []string{}),
		false,
		client,
		holder.New(),
	)
	dagRun.createPod()
	foundPod, err := dagRun.kubeClient.CoreV1().Pods(
		dagRun.Config.Namespace,
	).Get(
		context.TODO(),
		dagRun.pod.Name,
		k8sapi.GetOptions{},
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

func TestRunPod(t *testing.T) {
	// Test with logs and without logs
	client := fake.NewSimpleClientset()
	defer podutils.CleanUpEnvironment(client)
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
			expectedLogMessage := "Hello World!!!"
			dagRun := NewDAGRun(
				getTestDate(),
				getTestDAGConfig(
					"test-start-pod"+podutils.CleanK8sName(table.name),
					[]string{"echo", expectedLogMessage},
				),
				table.withLogs,
				client,
				holder.New(),
			)
			dagRun.Run()

			dagRun.holder.GetChannelGroup(dagRun.pod.Name).Ready <- dagRun.pod

			podCopy := dagRun.pod.DeepCopy()
			podCopy.Status.Phase = core.PodSucceeded
			dagRun.holder.GetChannelGroup(dagRun.pod.Name).Update <- podCopy

			dagRun.watcher.WaitForMonitorDone()

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
				logMsg := <-dagRun.Logs()
				logMsg = strings.ReplaceAll(logMsg, "\n", "")
				if logMsg != expectedLogMessage && logMsg != "fake logs" {
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
	client := fake.NewSimpleClientset()
	defer podutils.CleanUpEnvironment(client)
	dagRun := NewDAGRun(
		getTestDate(),
		getTestDAGConfig("test-delete-pod", []string{}),
		false,
		client,
		holder.New(),
	)
	podFrame := dagRun.getPodFrame()
	podsClient := client.CoreV1().Pods(dagRun.Config.Namespace)

	createdPod, err := podsClient.Create(context.TODO(), &podFrame, k8sapi.CreateOptions{})
	dagRun.pod = createdPod
	if err != nil {
		panic(err)
	}
	dagRun.DeletePod()
	list, err := podsClient.List(context.TODO(), k8sapi.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, pod := range list.Items {
		if jsonpanic.JSONPanic(*createdPod) == jsonpanic.JSONPanic(pod) {
			t.Errorf("Pod %s should have been deleted", createdPod.Name)
		}
	}
}

func TestStart(t *testing.T) {
	// Test with logs and without logs
	client := fake.NewSimpleClientset()
	defer podutils.CleanUpEnvironment(client)
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
			expectedLogMessage := "Hello World!!!"
			dagRun := NewDAGRun(
				getTestDate(),
				getTestDAGConfig(
					"test-start-pod"+podutils.CleanK8sName(table.name),
					[]string{"echo", expectedLogMessage},
				),
				table.withLogs,
				client,
				holder.New(),
			)
			go dagRun.Start()

			for {
				if dagRun.holder.Contains(dagRun.Name) {
					break
				}
				time.Sleep(1 * time.Millisecond)
			}

			dagRun.holder.GetChannelGroup(dagRun.Name).Ready <- dagRun.pod

			podCopy := dagRun.pod.DeepCopy()
			podCopy.Status.Phase = core.PodSucceeded
			dagRun.holder.GetChannelGroup(dagRun.Name).Update <- podCopy

			time.Sleep(1 * time.Millisecond)

			podList, err := client.CoreV1().Pods(
				dagRun.Config.Namespace,
			).List(
				context.TODO(),
				k8sapi.ListOptions{},
			)
			if err != nil {
				panic(err)
			}
			for _, item := range podList.Items {
				t.Log(item.Name)
				if item.Name == dagRun.Name {
					t.Errorf("Pod with name %s should have been deleted", dagRun.Name)
				}
			}
		}()

	}

}
