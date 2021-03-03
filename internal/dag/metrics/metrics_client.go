package metrics

import (
	"context"
	"encoding/binary"
	"fmt"
	"goflow/internal/jsonpanic"
	"goflow/internal/logs"
	"math/rand"
	"time"

	restclient "k8s.io/client-go/rest"

	"os"

	core "k8s.io/api/core/v1"
	k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// DAGMetricsClient handles all interactions with DAG metrics
type DAGMetricsClient struct {
	kubeClient kubernetes.Interface
	restConfig *restclient.Config
	testMode   bool
}

// PodMetrics holds information about the resource usage of a given pod
type PodMetrics struct {
	PodName string
	Time    time.Time
	Memory  uint32
	CPU     uint32
}

func (metric PodMetrics) String() string {
	return jsonpanic.JSONPanicFormat(metric)
}

func newPodMetrics(podName string) PodMetrics {
	return PodMetrics{podName, time.Now(), 0, 0}
}

// NewDAGMetricsClient returns a new DAGMetricsClient from a metrics clientset
func NewDAGMetricsClient(clientSet kubernetes.Interface, testMode bool) *DAGMetricsClient {
	return &DAGMetricsClient{clientSet, getRestConfig(), testMode}
}

type getMetricsOptions struct {
	kubeClient                      kubernetes.Interface
	restConfig                      *restclient.Config
	podName, command, containerName string
	testMode                        bool
}

func getContainerOutput(options getMetricsOptions) ([]byte, error) {
	reader := newWriteWrapper()
	err := execCmd(
		options.kubeClient,
		options.restConfig,
		options.podName,
		options.command,
		os.Stdin,
		&reader,
		os.Stderr,
		options.containerName,
	)
	if err != nil {
		return make([]byte, 0), err
	}
	return reader.data, nil
}

func getContainerIntMetric(options getMetricsOptions) (uint32, error) {
	if options.testMode {
		min := int32(1000000)
		return uint32(rand.Int31n(min*10)) + uint32(min), nil
	}
	data, err := getContainerOutput(options)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(data), nil
}

// getContainerMemory returns the container's current memory usage in bytes
func getContainerMemory(options getMetricsOptions) (uint32, error) {
	options.command = "cat /sys/fs/cgroup/memory/memory.usage_in_bytes"
	return getContainerIntMetric(options)
}

// getContainerCPU returns the container's current cpu usage in bytes
func getContainerCPU(options getMetricsOptions) (uint32, error) {
	options.command = "cat /sys/fs/cgroup/cpuacct/cpuacct.usage"
	return getContainerIntMetric(options)
}

// GetPodMetrics returns the total usage metrics for the containers in a given pod
func (client *DAGMetricsClient) GetPodMetrics(
	pod core.Pod,
) (PodMetrics, error) {
	metrics := newPodMetrics(pod.Name)
	hasActiveContainers := false
	for _, containerStatus := range pod.Status.ContainerStatuses {
		containerStarted := *containerStatus.Started
		if !containerStarted {
			continue
		}
		hasActiveContainers = true
		options := getMetricsOptions{
			kubeClient:    client.kubeClient,
			podName:       pod.Name,
			containerName: containerStatus.Name,
			restConfig:    client.restConfig,
			testMode:      client.testMode,
		}

		memory, err := getContainerMemory(options)
		if err != nil {
			logs.WarningLogger.Println("Error retrieving memory from container", err)
			continue
		}
		cpu, err := getContainerCPU(options)
		if err != nil {
			logs.WarningLogger.Println("Error retrieving CPU from container", err)
			continue
		}
		fmt.Println("Memory:", memory)
		fmt.Println("CPU:", cpu)
		metrics.Memory += memory
		metrics.CPU += cpu
	}
	if hasActiveContainers {
		return metrics, nil
	}
	return PodMetrics{}, fmt.Errorf("No available containers")
}

// ListPodMetrics returns a list of all metrics for pods in a given namespace
func (client *DAGMetricsClient) ListPodMetrics(namespace string) []PodMetrics {
	metricList := make([]PodMetrics, 0)
	pods, err := client.kubeClient.CoreV1().Pods(
		namespace,
	).List(
		context.TODO(),
		k8sapi.ListOptions{},
	)
	if err != nil {
		panic(err)
	}
	for _, pod := range pods.Items {
		metrics, err := client.GetPodMetrics(pod)
		if err == nil {
			metricList = append(metricList, metrics)
		}
	}
	return metricList
}
