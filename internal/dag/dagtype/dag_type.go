package dagtype

import (
	"encoding/json"
	"fmt"
	goflowconfig "goflow/internal/config"
	"goflow/internal/dag/activeruns"
	dagconfig "goflow/internal/dag/config"
	dagrun "goflow/internal/dag/run"
	"goflow/internal/k8s/pod/event/holder"
	"goflow/internal/logs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	dagtable "goflow/internal/dag/sql/dag"

	"github.com/robfig/cron"
	"k8s.io/client-go/kubernetes"
)

// ScheduleCache is a map from string to cron schedule
type ScheduleCache map[string]cron.Schedule

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
	timeLock            *sync.Mutex
	schedules           ScheduleCache
	*dagtable.TableClient
	filePath string
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
func CreateDAG(
	config *dagconfig.DAGConfig,
	code string,
	client kubernetes.Interface,
	schedules ScheduleCache,
	tableClient *dagtable.TableClient,
	filePath string,
) DAG {
	if config.Annotations == nil {
		config.Annotations = make(map[string]string)
	}
	if config.Labels == nil {
		config.Labels = make(map[string]string)
	}
	dag := DAG{
		Config:      config,
		Code:        code,
		DAGRuns:     make([]*dagrun.DAGRun, 0),
		kubeClient:  client,
		ActiveRuns:  activeruns.New(),
		timeLock:    &sync.Mutex{},
		schedules:   schedules,
		TableClient: tableClient,
		filePath:    filePath,
	}
	dag.StartDateTime = getDateFromString(dag.Config.StartDateTime)
	if dag.Config.EndDateTime != "" {
		dag.EndDateTime = getDateFromString(dag.Config.EndDateTime)
	}
	if dag.Config.MaxActiveRuns < 1 {
		panic("MaxActiveRuns must be greater than 0!")
	}
	dag.UpsertDag(
		dagtable.NewRow(
			0,
			dag.Config.Name,
			dag.Config.Namespace,
			"dag.Config.Version",
			dag.filePath,
			path.Ext(dag.filePath),
		),
	)
	return dag
}

func createDAGFromJSONBytes(
	dagBytes []byte,
	client kubernetes.Interface,
	goflowConfig goflowconfig.GoFlowConfig,
	scheduleCache ScheduleCache,
	tableClient *dagtable.TableClient,
	filePath string,
) (DAG, error) {
	dagConfigStruct := dagconfig.DAGConfig{}
	err := json.Unmarshal(dagBytes, &dagConfigStruct)
	dagConfigStruct.SetDefaults(goflowConfig)
	if err != nil {
		return DAG{}, err
	}

	// Validate schedule
	_, err = cron.Parse(dagConfigStruct.Schedule)
	if err != nil {
		panic(
			fmt.Sprintf(
				"DAG %s has a scheduling err with schedule \"%s\": ",
				dagConfigStruct.Name,
				dagConfigStruct.Schedule,
			) + err.Error(),
		)
	}

	dag := CreateDAG(
		&dagConfigStruct,
		string(dagBytes),
		client,
		scheduleCache,
		tableClient,
		filePath,
	)
	return dag, nil
}

// getDAGFromJSON creates a new dag struct from a dag file
func getDAGFromJSON(
	dagFilePath string,
	client kubernetes.Interface,
	goflowConfig goflowconfig.GoFlowConfig,
	scheduleCache ScheduleCache,
	tableClient *dagtable.TableClient,
) (DAG, error) {
	dagBytes, err := readDAGFile(dagFilePath)
	if err != nil {
		return DAG{}, err
	}
	dagJSON, err := createDAGFromJSONBytes(
		dagBytes,
		client,
		goflowConfig,
		scheduleCache,
		tableClient,
		dagFilePath,
	)
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
	schedules ScheduleCache,
	tableClient *dagtable.TableClient,
) []*DAG {
	files := getDirSliceRecur(folder)
	dags := make([]*DAG, 0, len(files))
	for _, file := range files {
		if strings.ToLower(filepath.Ext(file)) == ".json" {
			dag, err := getDAGFromJSON(file, client, goflowConfig, schedules, tableClient)
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

// getSchedule parses and caches or returns the stored schedule
func (dag *DAG) getSchedule() cron.Schedule {
	schedule, ok := dag.schedules[dag.Config.Schedule]
	if ok {
		return schedule
	}
	schedule, _ = cron.Parse(dag.Config.Schedule)
	dag.schedules[dag.Config.Schedule] = schedule
	return schedule
}

// getNextTime returns the next time according to the cron schedule
func (dag *DAG) getNextTime(lastTime time.Time) time.Time {
	schedule := dag.getSchedule()
	next := schedule.Next(lastTime)
	logs.InfoLogger.Println("Next:", next)
	return next
}

// AddNextDagRunIfReady adds the next dag run if ready for it, returns true if added, else false
func (dag *DAG) AddNextDagRunIfReady(holder *holder.ChannelHolder) (ready bool) {
	ready = dag.Ready()
	if ready {
		dag.timeLock.Lock()
		switch {
		case dag.MostRecentExecution.IsZero():
			dag.MostRecentExecution = dag.StartDateTime
		default:
			dag.MostRecentExecution = dag.getNextTime(dag.MostRecentExecution)
		}
		dagRun := dag.AddDagRun(dag.MostRecentExecution, dag.Config.WithLogs, holder)
		dag.timeLock.Unlock()
		dag.ActiveRuns.Inc()
		go dagRun.Start()
	}
	return
}

// TerminateAndDeleteRuns removes all active DAG runs and their associated pods
func (dag *DAG) TerminateAndDeleteRuns() {
	for _, run := range dag.DAGRuns {
		fmt.Println(run.MostRecentPod())
		run.DeletePod()
	}
}

// Ready returns true if the DAG is ready for another DAG Run to be created
func (dag *DAG) Ready() bool {
	currentTime := time.Now()
	scheduleReady := (dag.MostRecentExecution.Before(currentTime) ||
		dag.MostRecentExecution.Equal(currentTime) && dag.MostRecentExecution.Before(dag.EndDateTime))
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
