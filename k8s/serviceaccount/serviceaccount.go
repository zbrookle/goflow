package serviceaccount

import (
	"context"
	"goflow/k8s/pod/utils"

	core "k8s.io/api/core/v1"
	k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// Handler takes care of operations on services accounts of a given name and namespace
type Handler struct {
	name      string
	namespace string
	client    kubernetes.Interface
}

// New returns a new service account handler
func New(name string, namespace string, client kubernetes.Interface) Handler {
	name = utils.CleanK8sName(name)
	return Handler{name, namespace, client}
}

func (handler *Handler) serviceAccountClient() v1.ServiceAccountInterface {
	return handler.client.CoreV1().ServiceAccounts(handler.namespace)
}

func (handler *Handler) labels() map[string]string {
	return map[string]string{
		"App":  "goflow",
		"Name": handler.name,
	}
}

// Create creates a service account if it does not already exist
func (handler *Handler) Create() *core.ServiceAccount {
	serviceAccount, err := handler.serviceAccountClient().Create(
		context.TODO(),
		&core.ServiceAccount{
			TypeMeta: k8sapi.TypeMeta{
				Kind:       "ServiceAccount",
				APIVersion: "V1",
			},
			ObjectMeta: k8sapi.ObjectMeta{
				Name:      handler.name,
				Namespace: handler.namespace,
				Labels:    handler.labels(),
			},
		},
		k8sapi.CreateOptions{},
	)
	if err != nil {
		panic(err)
	}
	return serviceAccount
}

// Exists returns true if the given service account exists
func (handler *Handler) Exists() bool {
	serviceAccountList, err := handler.serviceAccountClient().List(
		context.TODO(),
		k8sapi.ListOptions{
			LabelSelector: utils.LabelSelectorString(handler.labels()),
		},
	)
	if err != nil {
		panic(err)
	}
	return len(serviceAccountList.Items) != 0
}
