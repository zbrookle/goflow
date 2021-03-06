package utils

import (
	"context"
	"goflow/internal/logs"
	"strings"

	core "k8s.io/api/core/v1"
	k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// AppSelectorKey is the key to use for selecting an application
const AppSelectorKey = "App"

// AppName is the name of the application
const AppName = "goflow"

func getAppLabelSelectorString() string {
	return LabelSelectorString(map[string]string{AppSelectorKey: AppName})
}

func getNamespaces(client kubernetes.Interface) []string {
	namespaceClient := client.CoreV1().Namespaces()
	namespaceList, err := namespaceClient.List(context.TODO(), k8sapi.ListOptions{})
	if err != nil {
		panic(err)
	}
	namespaceNames := make([]string, 0)
	for _, item := range namespaceList.Items {
		name := strings.TrimSpace(item.Name)
		if name != "" {
			namespaceNames = append(namespaceNames, name)
		}
	}
	return namespaceNames
}

// CleanUpPods deletes all pods currently present in the k8s cluster in all namespaces that are accessible
func CleanUpPods(client kubernetes.Interface) {
	namespaces := getNamespaces(client)
	for _, namespace := range namespaces {
		podsClient := client.CoreV1().Pods(namespace)
		podList, err := podsClient.List(
			context.TODO(),
			k8sapi.ListOptions{LabelSelector: getAppLabelSelectorString()},
		)
		if err != nil {
			panic(err)
		}
		for _, pod := range podList.Items {
			logs.InfoLogger.Printf(
				"Deleting pod \"%s\" in namespace \"%s\"\n",
				pod.Name,
				pod.Namespace,
			)
			podsClient.Delete(context.TODO(), pod.ObjectMeta.Name, k8sapi.DeleteOptions{})
		}
	}
}

// CleanUpServiceAccounts delete all associated application service accounts
func CleanUpServiceAccounts(client kubernetes.Interface) {
	namespaces := getNamespaces(client)
	for _, namespace := range namespaces {
		serviceAccountClient := client.CoreV1().ServiceAccounts(namespace)
		serviceAccountList, err := serviceAccountClient.List(
			context.TODO(),
			k8sapi.ListOptions{LabelSelector: getAppLabelSelectorString()},
		)
		if err != nil {
			panic(err)
		}
		for _, account := range serviceAccountList.Items {
			logs.InfoLogger.Printf(
				"Deleting service account \"%s\" in namespace \"%s\"\n",
				account.Name,
				namespace,
			)
			serviceAccountClient.Delete(context.TODO(), account.Name, k8sapi.DeleteOptions{})
		}
	}
}

// CleanUpEnvironment deletes all associated application resources
func CleanUpEnvironment(client kubernetes.Interface) {
	logs.InfoLogger.Println("Cleaning up...")
	CleanUpPods(client)
	CleanUpServiceAccounts(client)
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

// CleanK8sName returns a string with k8s incompatible characters removed
func CleanK8sName(name string) string {
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ReplaceAll(name, ":", "-")
	name = strings.ReplaceAll(name, " ", "")
	name = strings.ReplaceAll(name, "+", "plus")
	name = strings.ToLower(name)
	return name
}
