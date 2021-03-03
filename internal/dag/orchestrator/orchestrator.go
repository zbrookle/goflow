package orchestrator

import (
	"context"
	"fmt"
	dagconfig "goflow/internal/dag/config"
	dagtype "goflow/internal/dag/dagtype"
	"goflow/internal/dag/metrics"
	dagrun "goflow/internal/dag/run"
	"goflow/internal/database"
	"goflow/internal/jsonpanic"
	"goflow/internal/logs"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"goflow/internal/config"
	dagtable "goflow/internal/dag/sql/dag"
	dagruntable "goflow/internal/dag/sql/dagrun"
	k8sclient "goflow/internal/k8s/client"
	"goflow/internal/k8s/pod/event/holder"
	"goflow/internal/k8s/pod/inform"
	"goflow/internal/k8s/pod/utils"
	"goflow/internal/k8s/serviceaccount"

	"path"

	"github.com/kennygrant/sanitize"
	k8sapi "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Orchestrator holds information for all DAGs
type Orchestrator struct {
	dagMapLock        *sync.RWMutex
	dagMap            map[string]*dagtype.DAG
	kubeClient        kubernetes.Interface
	config            *config.GoFlowConfig
	channelHolder     *holder.ChannelHolder
	schedules         dagtype.ScheduleCache
	closingChannel    chan struct{}
	dagTableClient    *dagtable.TableClient
	dagrunTableClient *dagruntable.TableClient
	metricsClient     *metrics.DAGMetricsClient
}

// NewOrchestratorFromClientsAndConfig creates an orchestractor from a given k8s client and goflow config
func NewOrchestratorFromClientsAndConfig(
	client kubernetes.Interface,
	config *config.GoFlowConfig,
	metricsClient *metrics.DAGMetricsClient,
) *Orchestrator {
	sqlClient := database.NewSQLiteClient(config.DatabaseDNS)
	return &Orchestrator{
		&sync.RWMutex{},
		make(map[string]*dagtype.DAG),
		client,
		config,
		holder.New(),
		make(dagtype.ScheduleCache),
		make(chan struct{}),
		dagtable.NewTableClient(sqlClient),
		dagruntable.NewTableClient(sqlClient),
		metricsClient,
	}
}

// NewOrchestrator creates an empty instance of Orchestrator
func NewOrchestrator(configPath string) *Orchestrator {
	kubeClient := k8sclient.CreateKubeClient()
	return NewOrchestratorFromClientsAndConfig(
		kubeClient,
		config.CreateConfig(configPath),
		metrics.NewDAGMetricsClient(kubeClient, false),
	)
}

// AddDAG adds a DAG to the Orchestrator
func (orchestrator *Orchestrator) AddDAG(dag *dagtype.DAG) {
	logs.InfoLogger.Printf(
		"Added DAG '%s' which will run in namespace '%s', with configuration: %s",
		dag.Config.Name,
		dag.Config.Namespace,
		jsonpanic.JSONPanicFormat(dag.Config),
	)
	dag.LastUpdated = time.Now()
	orchestrator.dagMapLock.Lock()
	orchestrator.dagMap[dag.Config.Name] = dag
	orchestrator.dagMapLock.Unlock()
}

// UpdateDag updates a DAG to the most recently found configuration
func (orchestrator *Orchestrator) UpdateDag(dag *dagtype.DAG) {
	orchestrator.dagMapLock.Lock()
	dagRef := orchestrator.dagMap[dag.Config.Name]
	dagRef.LastUpdated = time.Now()
	logs.InfoLogger.Printf(
		"Updating DAG '%s' from namespace '%s', from configuration %s to configuration %s",
		dag.Config.Name,
		dag.Config.Namespace,
		dagRef.Config,
		dag.Config,
	)
	dagRef.Config = dag.Config
	orchestrator.dagMapLock.Unlock()
}

func (orchestrator *Orchestrator) addDAGServiceAccount(dag *dagtype.DAG) {
	serviceAccountHandler := serviceaccount.New(
		utils.AppName,
		dag.Config.Namespace,
		orchestrator.kubeClient,
	)
	if serviceAccountHandler.Exists() {
		return
	}
	serviceAccountHandler.Create()
}

// DeleteDAG removes a DAG from the orchestrator
func (orchestrator *Orchestrator) DeleteDAG(dagName string, namespace string) {
	dag := orchestrator.dagMap[dagName]
	dag.TerminateAndDeleteRuns()
	delete(orchestrator.dagMap, dagName)
}

// DAGs returns []DAGs with all DAGs present in the map
func (orchestrator Orchestrator) DAGs() dagtype.DAGList {
	dagSlice := make([]*dagtype.DAG, 0, len(orchestrator.dagMap))
	orchestrator.dagMapLock.RLock()
	for dagName := range orchestrator.dagMap {
		dagSlice = append(dagSlice, orchestrator.dagMap[dagName])
	}
	orchestrator.dagMapLock.RUnlock()
	return dagSlice
}

// isDagPresent returns true if the given dag is present
func (orchestrator Orchestrator) isDagPresent(dag dagtype.DAG) bool {
	_, ok := orchestrator.dagMap[dag.Config.Name]
	return ok
}

// isStoredDagDifferent returns true if the given dag source code is different
func (orchestrator Orchestrator) isStoredDagDifferent(dag dagtype.DAG) bool {
	currentDag, _ := orchestrator.dagMap[dag.Config.Name]
	return currentDag.Code != dag.Code
}

// GetDag returns the DAG with the given name
func (orchestrator Orchestrator) GetDag(dagName string) *dagtype.DAG {
	dag, _ := orchestrator.dagMap[dagName]
	return dag
}

// DagRuns returns all the dag runs across all dags
func (orchestrator Orchestrator) DagRuns() []dagrun.DAGRun {
	runs := make([]dagrun.DAGRun, 0)
	for _, dag := range orchestrator.DAGs() {
		for _, run := range dag.DAGRuns {
			runs = append(runs, *run)
		}
	}
	return runs
}

func (orchestrator *Orchestrator) collectDAG(dag *dagtype.DAG) {
	dagPresent := orchestrator.isDagPresent(*dag)
	if !dagPresent {
		orchestrator.addDAGServiceAccount(dag)
		orchestrator.AddDAG(dag)
	} else if dagPresent && orchestrator.isStoredDagDifferent(*dag) {
		logs.InfoLogger.Printf("Updating DAG %s which will run in namespace %s", dag.Config.Name, dag.Config.Namespace)
		logs.InfoLogger.Printf("Old DAG code: %s\n", orchestrator.GetDag(dag.Config.Name).Code)
		logs.InfoLogger.Printf("New DAG code: %s\n", dag.Code)
		orchestrator.UpdateDag(dag)
	}
}

// CollectDAGs fills up the dag map with existing dags
func (orchestrator *Orchestrator) CollectDAGs() {
	dagSlice := dagtype.GetDAGSFromFolder(
		orchestrator.config.DAGPath,
		orchestrator.kubeClient,
		orchestrator.metricsClient,
		*orchestrator.config,
		orchestrator.schedules,
		orchestrator.dagTableClient,
		orchestrator.dagrunTableClient,
	)
	for _, dag := range dagSlice {
		orchestrator.collectDAG(dag)
	}
}

// RunDags schedules pods for all dags that are ready
func (orchestrator *Orchestrator) RunDags() {
	for _, dag := range orchestrator.DAGs() {
		dag.AddNextDagRunIfReady(orchestrator.channelHolder)
	}
}

func cycleUntilChannelClose(
	callable func(),
	close chan struct{},
	cycleDuration time.Duration,
	loopName string,
) {
	for {
		select {
		case _, ok := <-close:
			if !ok {
				logs.InfoLogger.Printf("Closing %s\n", loopName)
				return
			}
		default:
			callable()
			time.Sleep(cycleDuration)
		}
	}
}

func (orchestrator *Orchestrator) getTaskInformer() inform.TaskInformer {
	return inform.New(orchestrator.kubeClient, orchestrator.channelHolder)
}

func (orchestrator *Orchestrator) setupDatabaseTables() {
	orchestrator.dagTableClient.CreateTable()
	orchestrator.dagrunTableClient.CreateTable()
}

// Start begins the orchestrator event loop
func (orchestrator *Orchestrator) Start(cycleDuration time.Duration) {
	orchestrator.setupDatabaseTables()
	taskInformer := orchestrator.getTaskInformer()
	taskInformer.Start()
	go cycleUntilChannelClose(
		orchestrator.CollectDAGs,
		orchestrator.closingChannel,
		cycleDuration,
		"Collect DAGs",
	)
	go cycleUntilChannelClose(
		orchestrator.RunDags,
		orchestrator.closingChannel,
		cycleDuration,
		"Run DAGs",
	)
	go orchestrator.StoreMetricsEvery(2)
}

// Wait blocks the current thread until the orchestrator has terminated
func (orchestrator *Orchestrator) Wait() {
	<-orchestrator.closingChannel
}

// Stop terminates the orchestrators cycles
func (orchestrator *Orchestrator) Stop() {
	close(orchestrator.closingChannel)
}

// WriteDAGFile writes a new DAG to the dag file location
func (orchestrator *Orchestrator) WriteDAGFile(config *dagconfig.DAGConfig) (int, error) {
	if !config.IsNameValid() {
		return http.StatusBadRequest, fmt.Errorf(
			"DAG name must match the pattern \"%s\"",
			config.Pattern(),
		)
	}
	cleanName := sanitize.Path(config.Name)
	pathToFile := path.Join(orchestrator.config.DAGPath, fmt.Sprintf("%s.json", cleanName))
	_, err := filepath.Rel(orchestrator.config.DAGPath, pathToFile)
	if err != nil {
		return http.StatusBadRequest, err
	}
	_, err = os.Stat(pathToFile)
	if os.IsExist(err) {
		return http.StatusConflict, fmt.Errorf("DAG with given name already present")
	}
	err = config.WriteToFile(pathToFile)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, err
}

// StoreMetricsEvery stores usage metrics for all pods every 'seconds' unit of time
func (orchestrator *Orchestrator) StoreMetricsEvery(seconds time.Duration) {
	for {
		namespaces, err := orchestrator.kubeClient.CoreV1().Namespaces().List(
			context.TODO(),
			k8sapi.ListOptions{},
		)
		if err != nil {
			panic(err)
		}
		for _, namespace := range namespaces.Items {
			if namespace.Name[:4] == "kube" {
				continue
			}
			fmt.Println(namespace.Name)
			fmt.Println("-------------")
			metrics := orchestrator.metricsClient.ListPodMetrics(namespace.Name)
			fmt.Println(metrics)
		}
		time.Sleep(seconds * time.Second)
	}
}
