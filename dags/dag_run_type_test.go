package dags

import (
	"context"
	"goflow/jsonpanic"
	"goflow/k8sclient"
	"goflow/testutils"
	"testing"

	// "time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreatePod(t *testing.T) {
	defer testutils.CleanUpPods(KUBECLIENT)
	dagRun := createDagRun(getTestDate(), getTestDAGFakeClient())
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
	realClient := k8sclient.CreateKubeClient()
	defer testutils.CleanUpPods(realClient)
	dagRun := createDagRun(getTestDate(), getTestDAGRealClient())
	podName := dagRun.Start()
	_, err := realClient.CoreV1().Pods(
		dagRun.DAG.Config.Namespace,
	).Get(
		context.TODO(),
		podName,
		v1.GetOptions{},
	)

	// time.Sleep(30 * time.Second)

	if err != nil {
		t.Errorf("Pod %s could not be found", podName)
		panic("Pod not found")
	}
	// if pod.Status. != 1 {
	// 	t.Errorf("Pod %s did not complete yet", podName)
	// }
}

func TestDeletePod(t *testing.T) {
	defer testutils.CleanUpPods(KUBECLIENT)
	dagRun := createDagRun(getTestDate(), getTestDAGFakeClient())
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
