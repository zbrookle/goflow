package dags

import (
	"context"
	"encoding/json"
	"goflow/logs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// DAG is directed acyclic graph for hold job information
type DAG struct {
	Name        string
	Namespace   string
	Schedule    string
	DockerImage string
	RetryPolicy string
	Command     string
	Parallelism int32
	TimeLimit   int64
	Retries     int32
	Labels      map[string]string
	Annotations map[string]string
	DAGRuns     []*DAGRun
	kubeClient  kubernetes.Interface
}

// DAGRun is a single run of a given dag - corresponds with a kubernetes Job
type DAGRun struct {
	Name          string
	DAG           *DAG
	ExecutionDate k8sapi.Time // This is the date that will be passed to the job that runs
	Start         k8sapi.Time
	End           k8sapi.Time
	Job           *batch.Job
}

func readDAGFile(dagFilePath string) []byte {
	dat, err := ioutil.ReadFile(dagFilePath)
	if err != nil {
		panic(err)
	}
	return dat
}

func createDAGFromJSONBytes(dagBytes []byte) (DAG, error) {
	dagStruct := DAG{}
	err := json.Unmarshal(dagBytes, &dagStruct)
	if err != nil {
		return dagStruct, err
	}
	dagStruct.DAGRuns = make([]*DAGRun, 0)
	return dagStruct, nil
}

// getDAGFromJSON creates a new dag struct from a dag file
func getDAGFromJSON(dagFilePath string) (DAG, error) {
	dagBytes := readDAGFile(dagFilePath)
	dagJSON, err := createDAGFromJSONBytes(dagBytes)
	if err != nil {
		logs.ErrorLogger.Printf("Error parsing dag file %s", dagFilePath)
		return DAG{}, err
	}
	return dagJSON, nil
}

// getDirSliceRecur recursively retrieves all file names from the directory
func getDirSliceRecur(directory string) []string {
	files := []string{}
	dagFileRegex := regexp.MustCompile(".*_dag.*\\.(go|json|py)")
	appendToFiles := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if dagFileRegex.Match([]byte(path)) {
			files = append(files, path)
		}
		return nil
	}
	err := filepath.Walk(directory, appendToFiles)
	if err != nil {
		panic(err)
	}
	return files
}

// GetDAGSFromFolder returns a slice of DAG structs, one for each DAG file
// Each file must have the "dag" suffix
// E.g., my_dag.py, some_dag.json
func GetDAGSFromFolder(folder string) []DAG {
	files := getDirSliceRecur(folder)
	dags := make([]DAG, 0, len(files))
	for _, file := range files {
		if strings.ToLower(filepath.Ext(file)) == ".json" {
			dag, err := getDAGFromJSON(file)
			if err == nil {
				dags = append(dags, dag)
			}
		}
	}
	return dags
}

// NewDAG creates a new dag initialized with an empty DAGRuns slice
func NewDAG(
	name string,
	namespace string,
	schedule string,
	dockerImage string,
	retryPolicy string,
	kubeClient kubernetes.Interface,
) *DAG {
	return &DAG{
		Name:        name,
		Namespace:   namespace,
		Schedule:    schedule,
		DockerImage: dockerImage,
		RetryPolicy: retryPolicy,
		DAGRuns:     make([]*DAGRun, 0),
		kubeClient:  kubeClient,
	}
}

func createDagRun(executionDate time.Time, dag *DAG) *DAGRun {
	return &DAGRun{
		Name: dag.Name + executionDate.String(),
		DAG:  dag,
		ExecutionDate: k8sapi.Time{
			Time: executionDate,
		},
		Start: k8sapi.Time{
			Time: time.Now(),
		},
		End: k8sapi.Time{
			Time: time.Time{},
		},
	}
}

// AddDagRun adds a DagRun for a scheduled point to the orchestrators set of dags
func (dag *DAG) AddDagRun(executionDate time.Time) {
	dagRun := createDagRun(executionDate, dag)
	dag.DAGRuns = append(dag.DAGRuns, dagRun)
}

// TerminateAndDeleteRuns removes all active DAG runs and their associated jobs
func (dag *DAG) TerminateAndDeleteRuns() {
	for _, run := range dag.DAGRuns {
		run.deleteJob()
	}
}

// getJobFrame returns a job from a DagRun
func (dagRun DAGRun) getJobFrame() batch.Job {
	dag := dagRun.DAG
	return batch.Job{
		TypeMeta: k8sapi.TypeMeta{
			Kind:       "Job",
			APIVersion: "v1",
		},
		ObjectMeta: k8sapi.ObjectMeta{
			Name:        dagRun.Name,
			Namespace:   dag.Namespace,
			Labels:      dag.Labels,
			Annotations: dag.Annotations,
		},
		Spec: batch.JobSpec{
			Parallelism:           &dag.Parallelism,
			ActiveDeadlineSeconds: &dag.TimeLimit,
			BackoffLimit:          &dag.Retries,
			Template: core.PodTemplateSpec{
				ObjectMeta: k8sapi.ObjectMeta{
					Name:      dag.Name,
					Namespace: dag.Namespace,
					// Labels: map[string]string{
					// 	"": "",
					// },
					// Annotations: map[string]string{
					// 	"": "",
					// },
				},
				Spec: core.PodSpec{
					Volumes:                       nil,
					Containers:                    nil,
					EphemeralContainers:           nil,
					RestartPolicy:                 "",
					TerminationGracePeriodSeconds: nil,
					ActiveDeadlineSeconds:         nil,
				},
			},
		},
	}
}

// CreateJob creates and registers a new job with
func (dagRun *DAGRun) CreateJob() {
	dag := dagRun.DAG
	jobFrame := dagRun.getJobFrame()
	job, err := dag.kubeClient.BatchV1().Jobs(
		dag.Namespace,
	).Create(
		context.TODO(),
		&jobFrame,
		k8sapi.CreateOptions{},
	)
	if err != nil {
		panic(err)
	}
	dagRun.Job = job
}

// deleteJob
func (dagRun *DAGRun) deleteJob() {
	err := dagRun.DAG.kubeClient.BatchV1().Jobs(
		dagRun.DAG.Namespace,
	).Delete(
		context.TODO(),
		dagRun.Name,
		k8sapi.DeleteOptions{},
	)
	if err != nil {
		panic(err)
	}
}
