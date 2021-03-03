package client

import (
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

func getConfigFromHome() *rest.Config {
	var kubeconfig string
	home := homedir.HomeDir()
	kubeconfig = filepath.Join(home, ".kube", "config")

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	return config
}

// CreateKubeClient returns a kubernetes client authenticated using kubeconfig
func CreateKubeClient() *kubernetes.Clientset {
	config := getConfigFromHome()
	return kubernetes.NewForConfigOrDie(config)
}

// CreateMetricsClient returns a new k8s metrics client
func CreateMetricsClient() *metrics.Clientset {
	config := getConfigFromHome()
	return metrics.NewForConfigOrDie(config)
}
