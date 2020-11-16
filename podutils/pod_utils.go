package podutils

import (
	"context"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CleanUpPods deletes all pods currently present in the k8s cluster in all namespaces that are accessible
func CleanUpPods(client kubernetes.Interface) {
	fmt.Println("Cleaning up...")
	namespaceClient := client.CoreV1().Namespaces()
	namespaceList, err := namespaceClient.List(context.TODO(), v1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, namespace := range namespaceList.Items {
		podsClient := client.CoreV1().Pods(namespace.Name)
		podList, err := podsClient.List(context.TODO(), v1.ListOptions{})
		if err != nil {
			panic(err)
		}
		for _, pod := range podList.Items {
			fmt.Printf("Deleting pod %s in namespace %s\n", pod.ObjectMeta.Name, namespace.Name)
			podsClient.Delete(context.TODO(), pod.ObjectMeta.Name, v1.DeleteOptions{})
		}
	}
}
