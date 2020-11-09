package testutils

import (
	"context"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CleanUpJobs deletes all jobs currently present in the k8s cluster in all namespaces that are accessible
func CleanUpJobs(client kubernetes.Interface) {
	fmt.Println("Cleaning up...")
	namespaceClient := client.CoreV1().Namespaces()
	namespaceList, err := namespaceClient.List(context.TODO(), v1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, namespace := range namespaceList.Items {
		jobsClient := client.BatchV1().Jobs(namespace.Name)
		jobList, err := jobsClient.List(context.TODO(), v1.ListOptions{})
		if err != nil {
			panic(err)
		}
		for _, job := range jobList.Items {
			fmt.Printf("Deleting job %s in namespace %s\n", job.ObjectMeta.Name, namespace.Name)
			jobsClient.Delete(context.TODO(), job.ObjectMeta.Name, v1.DeleteOptions{})
		}
	}
}
