package cron

import (
	batchclient "k8s.io/client-go/kubernetes/typed/batch/v1"
	rest "k8s.io/client-go/rest"
	"os/exec"
	"strings"
)

func getMinikubeIP() string {
	bytesIP, err := exec.Command("minikube", "ip").Output()
	if err != nil {
		panic(err)
	}
	bytesString := string(bytesIP)
	return strings.TrimSpace(bytesString)
}

// CreateKubeBatchClient returns a kubernetes client authenticated using kubeconfig
func CreateKubeBatchClient() *batchclient.BatchV1Client {
	config := &rest.Config{Host: getMinikubeIP()}

	return batchclient.NewForConfigOrDie(config)
}
