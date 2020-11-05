package cron

import (
	rest "k8s.io/client-go/rest"
	"os/exec"
	"strings"
	"k8s.io/client-go/kubernetes"
	"github.com/davecgh/go-spew/spew"
)

func getMinikubeIP() string {
	bytesIP, err := exec.Command("minikube", "ip").Output()
	if err != nil {
		panic(err)
	}
	bytesString := string(bytesIP)
	return strings.TrimSpace(bytesString)
}

// CreateKubeClient returns a kubernetes client authenticated using kubeconfig
func CreateKubeClient() *kubernetes.Clientset {
	config := &rest.Config{Host: getMinikubeIP()}

	return kubernetes.NewForConfigOrDie(config)
}
