package dags

import (
	"encoding/json"
	"goflow/dagconfig"
	"goflow/dagrun"
	"goflow/logs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

// DAG is directed acyclic graph for hold job information
type DAG struct {
	Config              *dagconfig.DAGConfig
	Code                string
	StartDateTime       time.Time
	EndDateTime         time.Time
	DAGRuns             []*dagrun.DAGRun
	kubeClient          kubernetes.Interface
	ActiveRuns          int
	MostRecentExecution time.Time
}

func readDAGFile(dagFilePath string) []byte {
	dat, err := ioutil.ReadFile(dagFilePath)
	if err != nil {
		panic(err)
	}
	return dat
}

func getDateFromString(dateStr string) time.Time {
	time, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		panic(err)
	}
	return time
}

// CreateDAG returns a dag using the configuration passed and stores the code string
func CreateDAG(config *dagconfig.DAGConfig, code string, client kubernetes.Interface) DAG {
	if config.Annotations == nil {
		config.Annotations = make(map[string]string)
	}
	if config.Labels == nil {
		config.Labels = make(map[string]string)
	}
	dag := DAG{Config: config, Code: code, DAGRuns: make([]*dagrun.DAGRun, 0), kubeClient: client}
	dag.StartDateTime = getDateFromString(dag.Config.StartDateTime)
	if dag.Config.EndDateTime != "" {
		dag.EndDateTime = getDateFromString(dag.Config.EndDateTime)
	}
	if dag.Config.MaxActiveRuns < 1 {
		panic("MaxActiveRuns must be greater than 0!")
	}
	return dag
}

func createDAGFromJSONBytes(dagBytes []byte, client kubernetes.Interface) (DAG, error) {
	dagConfigStruct := dagconfig.DAGConfig{}
	err := json.Unmarshal(dagBytes, &dagConfigStruct)
	if err != nil {
		return DAG{}, err
	}
	dag := CreateDAG(&dagConfigStruct, string(dagBytes), client)
	return dag, nil
}

// getDAGFromJSON creates a new dag struct from a dag file
func getDAGFromJSON(dagFilePath string, client kubernetes.Interface) (DAG, error) {
	dagBytes := readDAGFile(dagFilePath)
	dagJSON, err := createDAGFromJSONBytes(dagBytes, client)
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
			dag, err := getDAGFromJSON(file, fake.NewSimpleClientset())
			if err == nil {
				dags = append(dags, &dag)
			}
		}
	}
	return dags
}

// AddDagRun adds a DagRun for a scheduled point to the orchestrators set of dags
func (dag *DAG) AddDagRun(executionDate time.Time, withLogs bool) *dagrun.DAGRun {
	dagRun := dagrun.NewDAGRun(executionDate, dag.Config, withLogs, dag.kubeClient)
	dag.DAGRuns = append(dag.DAGRuns, dagRun)
	dag.ActiveRuns++
	return dagRun
}

// AddNextDagRunIfReady adds the next dag run if ready for it
func (dag *DAG) AddNextDagRunIfReady() {
	if dag.Ready() {
		if dag.MostRecentExecution.IsZero() {
			dag.MostRecentExecution = dag.StartDateTime
		}
		// !!!! Bug still occurring here need to figure out why !!!!!
		// dagRun :=
		dag.AddDagRun(dag.MostRecentExecution, dag.Config.WithLogs)
		// go dagRun.Start()
	}
}

// TerminateAndDeleteRuns removes all active DAG runs and their associated pods
func (dag *DAG) TerminateAndDeleteRuns() {
	for _, run := range dag.DAGRuns {
		run.DeletePod()
	}
}

// Ready returns true if the DAG is ready for another DAG Run to be created
func (dag *DAG) Ready() bool {
	currentTime := time.Now()
	scheduleReady := dag.MostRecentExecution.Before(currentTime) ||
		dag.MostRecentExecution.Equal(currentTime)
	return (dag.ActiveRuns < dag.Config.MaxActiveRuns) && scheduleReady
}

// Marshal returns the JSON byte slice representation of the DAG
func (dag *DAG) Marshal() []byte {
	jsonString, err := json.Marshal(dag)
	if err != nil {
		panic(err)
	}
	return jsonString
}

// String returns a nice JSON representation of the dag
func (dag *DAG) String() string {
	jsonString, err := json.MarshalIndent(dag, "", "\t")
	if err != nil {
		panic(err)
	}
	return string(jsonString)
}
