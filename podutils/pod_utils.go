package podutils

import (
	"context"
	"fmt"

	k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// AppSelectorKey is the key to use for selecting an application
const AppSelectorKey = "App"

// AppName is the name of the application
const AppName = "goflow"

// CleanUpPods deletes all pods currently present in the k8s cluster in all namespaces that are accessible
func CleanUpPods(client kubernetes.Interface) {
	fmt.Println("Cleaning up...")
	namespaceClient := client.CoreV1().Namespaces()
	namespaceList, err := namespaceClient.List(context.TODO(), v1.ListOptions{})
	if err != nil {
		panic(err)
	}
	labelSelector := LabelSelectorString(map[string]string{AppSelectorKey: AppName})
	for _, namespace := range namespaceList.Items {
		podsClient := client.CoreV1().Pods(namespace.Name)
		podList, err := podsClient.List(
			context.TODO(),
			v1.ListOptions{LabelSelector: labelSelector},
		)
		if err != nil {
			panic(err)
		}
		for _, pod := range podList.Items {
			fmt.Printf("Deleting pod %s in namespace %s\n", pod.ObjectMeta.Name, namespace.Name)
			podsClient.Delete(context.TODO(), pod.ObjectMeta.Name, v1.DeleteOptions{})
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
