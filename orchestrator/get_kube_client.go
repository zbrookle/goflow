package orchestrator

import (
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// createKubeClient returns a kubernetes client authenticated using kubeconfig
func createKubeClient() *kubernetes.Clientset {
	var kubeconfig string
	home := homedir.HomeDir()
	kubeconfig = filepath.Join(home, ".kube", "config")

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	return kubernetes.NewForConfigOrDie(config)
}
