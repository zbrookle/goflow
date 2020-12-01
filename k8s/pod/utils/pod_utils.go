package utils

import (
	"context"
	"goflow/logs"

	core "k8s.io/api/core/v1"
	k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// AppSelectorKey is the key to use for selecting an application
const AppSelectorKey = "App"

// AppName is the name of the application
const AppName = "goflow"

// CleanUpPods deletes all pods currently present in the k8s cluster in all namespaces that are accessible
func CleanUpPods(client kubernetes.Interface) {
	logs.InfoLogger.Println("Cleaning up...")
	namespaceClient := client.CoreV1().Namespaces()
	namespaceList, err := namespaceClient.List(context.TODO(), k8sapi.ListOptions{})
	if err != nil {
		panic(err)
	}
	labelSelector := LabelSelectorString(map[string]string{AppSelectorKey: AppName})
	for _, namespace := range namespaceList.Items {
		podsClient := client.CoreV1().Pods(namespace.Name)
		podList, err := podsClient.List(
			context.TODO(),
			k8sapi.ListOptions{LabelSelector: labelSelector},
		)
		if err != nil {
			panic(err)
		}
		for _, pod := range podList.Items {
			logs.InfoLogger.Printf(
				"Deleting pod %s in namespace %s\n",
				pod.ObjectMeta.Name,
				namespace.Name,
			)
			podsClient.Delete(context.TODO(), pod.ObjectMeta.Name, k8sapi.DeleteOptions{})
		}
	}
}

// LabelSelectorString returns a label selector for pods based on a given string map
func LabelSelectorString(labelMap map[string]string) string {
	nameSelector, err := k8sapi.LabelSelectorAsSelector(&k8sapi.LabelSelector{
		MatchLabels: labelMap,
	})
	if err != nil {
		panic(err)
	}
	return nameSelector.String()
}

// CreateTestPod creates and returns a busybox pod using the given PodInterface and names. It performs the command
// echo "123 test"
func CreateTestPod(
	podsClient *v1.PodInterface,
	podName string,
	namespace string,
	phase core.PodPhase,
) *core.Pod {
	testPodFrame := core.Pod{
		TypeMeta: k8sapi.TypeMeta{
			Kind:       "Pod",
			APIVersion: "V1",
		},
		ObjectMeta: k8sapi.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Labels:    map[string]string{AppSelectorKey: AppName},
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
		Status: core.PodStatus{Phase: phase},
	}
	pod, err := (*podsClient).Create(context.TODO(), &testPodFrame, k8sapi.CreateOptions{})
	if err != nil {
		panic(err)
	}
	return pod
}
