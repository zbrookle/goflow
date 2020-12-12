package testutils

import (
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	fakecore "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	testcore "k8s.io/client-go/testing"
)

// GetPodClientWithTestWatcher returns a podClient along with a fake watcher for adding event watching for test pods
func GetPodClientWithTestWatcher(
	client kubernetes.Interface,
	namespace string,
) (*v1.PodInterface, *watch.FakeWatcher) {
	podsClient := client.CoreV1().Pods(namespace)
	fakePodClient := podsClient.(*fakecore.FakePods)
	watcher := watch.NewFake()
	fakePodClient.Fake.PrependWatchReactor(
		"pods",
		testcore.DefaultWatchReactor(watcher, nil),
	)
	return &podsClient, watcher
}
