package dags

import (
	"context"
	"encoding/json"
	"fmt"
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

// DAGConfig is a struct storing the configurable values provided from the user in the DAG
// definition file
type DAGConfig struct {
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
}

// Marshal returns a json bytes representation of DAGConfig
func (config DAGConfig) Marshal() []byte {
	jsonBytes, err := json.Marshal(config)
	if err != nil {
		panic(err)
	}
	return jsonBytes
}

// JSON returns a json string representation of DAGConfig
func (config DAGConfig) JSON() string {
	return string(config.Marshal())
}

// DAG is directed acyclic graph for hold job information
type DAG struct {
	Config     *DAGConfig
	Code       string
	DAGRuns    []*DAGRun
	kubeClient kubernetes.Interface
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

// // Config returns a string Config representation of a DAGs configurable values
// func (dag DAG) Config() string {
// 	return dag.dagConfig
// }

func readDAGFile(dagFilePath string) []byte {
	dat, err := ioutil.ReadFile(dagFilePath)
	if err != nil {
		panic(err)
	}
	return dat
}

func createDAGFromDagConfigAndCode(config *DAGConfig, code string) DAG {
	dag := DAG{Config: config}
	dag.DAGRuns = make([]*DAGRun, 0)
	dag.Code = code
	return dag
}

func createDAGFromJSONBytes(dagBytes []byte) (DAG, error) {
	dagConfigStruct := DAGConfig{}
	fmt.Println(string(dagBytes))
	err := json.Unmarshal(dagBytes, &dagConfigStruct)
	if err != nil {
		return DAG{}, err
	}
	dag := createDAGFromDagConfigAndCode(&dagConfigStruct, string(dagBytes))
	return dag, nil
}

// getDAGFromJSON creates a new dag struct from a dag file
func getDAGFromJSON(dagFilePath string) (DAG, error) {
	dagBytes := readDAGFile(dagFilePath)
	dagJSON, err := createDAGFromJSONBytes(dagBytes)
	if err != nil {
		logs.ErrorLogger.Printf("Error parsing dag file %s", dagFilePath)
		return DAG{}, err
	}
	dagJSON.Code = string(dagBytes)
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
		logs.ErrorLogger.Println(directory, "not found")
		panic(err)
	}
	return files
}

// GetDAGSFromFolder returns a slice of DAG structs, one for each DAG file
// Each file must have the "dag" suffix
// E.g., my_dag.py, some_dag.json
func GetDAGSFromFolder(folder string) []*DAG {
	files := getDirSliceRecur(folder)
	dags := make([]*DAG, 0, len(files))
	for _, file := range files {
		if strings.ToLower(filepath.Ext(file)) == ".json" {
			dag, err := getDAGFromJSON(file)
			if err == nil {
				dags = append(dags, &dag)
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
		Config: &DAGConfig{Name: name,
			Namespace:   namespace,
			Schedule:    schedule,
			DockerImage: dockerImage,
			RetryPolicy: retryPolicy},
		DAGRuns:    make([]*DAGRun, 0),
		kubeClient: kubeClient,
	}
}

func createDagRun(executionDate time.Time, dag *DAG) *DAGRun {
	return &DAGRun{
		Name: dag.Config.Name + executionDate.String(),
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
			Namespace:   dag.Config.Namespace,
			Labels:      dag.Config.Labels,
			Annotations: dag.Config.Annotations,
		},
		Spec: batch.JobSpec{
			Parallelism:           &dag.Config.Parallelism,
			ActiveDeadlineSeconds: &dag.Config.TimeLimit,
			BackoffLimit:          &dag.Config.Retries,
			Template: core.PodTemplateSpec{
				ObjectMeta: k8sapi.ObjectMeta{
					Name:      dag.Config.Name,
					Namespace: dag.Config.Namespace,
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
		dag.Config.Namespace,
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
		dagRun.DAG.Config.Namespace,
	).Delete(
		context.TODO(),
		dagRun.Name,
		k8sapi.DeleteOptions{},
	)
	if err != nil {
		panic(err)
	}
}
