package dagtype

import (
	"encoding/json"
	goflowconfig "goflow/config"
	"goflow/dag/activeruns"
	dagconfig "goflow/dag/config"
	dagrun "goflow/dag/run"
	"goflow/k8s/pod/event/holder"
	"goflow/logs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
)

// DAG is directed acyclic graph for hold job information
type DAG struct {
	Config              *dagconfig.DAGConfig
	Code                string
	StartDateTime       time.Time
	EndDateTime         time.Time
	DAGRuns             []*dagrun.DAGRun
	kubeClient          kubernetes.Interface
	ActiveRuns          *activeruns.ActiveRuns
	MostRecentExecution time.Time
}

func readDAGFile(dagFilePath string) ([]byte, error) {
	dat, err := ioutil.ReadFile(dagFilePath)
	if err != nil {
		return nil, err
	}
	return dat, nil
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
	dag := DAG{
		Config:     config,
		Code:       code,
		DAGRuns:    make([]*dagrun.DAGRun, 0),
		kubeClient: client,
		ActiveRuns: activeruns.New(),
	}
	dag.StartDateTime = getDateFromString(dag.Config.StartDateTime)
	if dag.Config.EndDateTime != "" {
		dag.EndDateTime = getDateFromString(dag.Config.EndDateTime)
	}
	if dag.Config.MaxActiveRuns < 1 {
		panic("MaxActiveRuns must be greater than 0!")
	}
	return dag
}

func createDAGFromJSONBytes(
	dagBytes []byte,
	client kubernetes.Interface,
	goflowConfig goflowconfig.GoFlowConfig,
) (DAG, error) {
	dagConfigStruct := dagconfig.DAGConfig{}
	err := json.Unmarshal(dagBytes, &dagConfigStruct)
	dagConfigStruct.SetDefaults(goflowConfig)
	if err != nil {
		return DAG{}, err
	}
	dag := CreateDAG(&dagConfigStruct, string(dagBytes), client)
	return dag, nil
}

// getDAGFromJSON creates a new dag struct from a dag file
func getDAGFromJSON(
	dagFilePath string,
	client kubernetes.Interface,
	goflowConfig goflowconfig.GoFlowConfig,
) (DAG, error) {
	dagBytes, err := readDAGFile(dagFilePath)
	if err != nil {
		return DAG{}, err
	}
	dagJSON, err := createDAGFromJSONBytes(dagBytes, client, goflowConfig)
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
	if os.IsNotExist(err) {
		logs.WarningLogger.Printf("Directory \"%s\" not found", directory)
		return files
	}
	if err != nil {
		logs.ErrorLogger.Println(err)
		panic(err)
	}
	return files
}

// GetDAGSFromFolder returns a slice of DAG structs, one for each DAG file
// Each file must have the "dag" suffix
// E.g., my_dag.py, some_dag.json
func GetDAGSFromFolder(
	folder string,
	client kubernetes.Interface,
	goflowConfig goflowconfig.GoFlowConfig,
) []*DAG {
	files := getDirSliceRecur(folder)
	dags := make([]*DAG, 0, len(files))
	for _, file := range files {
		if strings.ToLower(filepath.Ext(file)) == ".json" {
			dag, err := getDAGFromJSON(file, client, goflowConfig)
			if os.ErrNotExist == err {
				logs.ErrorLogger.Printf("File %s no longer exists", file)
			}
			if err == nil {
				dags = append(dags, &dag)
			}
		}
	}
	return dags
}

// AddDagRun adds a DagRun for a scheduled point to the orchestrators set of dags
func (dag *DAG) AddDagRun(
	executionDate time.Time,
	withLogs bool,
	holder *holder.ChannelHolder,
) *dagrun.DAGRun {
	dagRun := dagrun.NewDAGRun(
		executionDate,
		dag.Config,
		withLogs,
		dag.kubeClient,
		holder,
		dag.ActiveRuns,
	)
	dag.DAGRuns = append(dag.DAGRuns, dagRun)
	return dagRun
}

// AddNextDagRunIfReady adds the next dag run if ready for it
func (dag *DAG) AddNextDagRunIfReady(holder *holder.ChannelHolder) {
	if dag.Ready() {
		if dag.MostRecentExecution.IsZero() {
			dag.MostRecentExecution = dag.StartDateTime
		}
		dagRun := dag.AddDagRun(dag.MostRecentExecution, dag.Config.WithLogs, holder)
		dag.ActiveRuns.Inc()
		go dagRun.Start()
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
	logs.InfoLogger.Printf("dag %s is ready: %v\n", dag.Config.Name, scheduleReady)
	return (dag.ActiveRuns.Get() < dag.Config.MaxActiveRuns) && scheduleReady
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
