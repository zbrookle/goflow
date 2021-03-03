package metrics

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
	restclient "k8s.io/client-go/rest"

	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

func getRestConfig() *restclient.Config {
	kubeConfigFlags := genericclioptions.NewConfigFlags(
		true,
	).WithDeprecatedPasswordFlag() // TODO: Figure out how to set up config without using kubectl
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	commandFactory := cmdutil.NewFactory(matchVersionKubeConfigFlags)
	restConfig, err := commandFactory.ToRESTConfig()
	if err != nil {
		panic(err)
	}
	return restConfig
}
