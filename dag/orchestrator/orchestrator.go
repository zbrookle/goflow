package orchestrator

import (
	dagtype "goflow/dag/dagtype"
	dagrun "goflow/dag/run"
	"goflow/jsonpanic"
	"goflow/logs"
	"sync"
	"time"

	"goflow/config"
	k8sclient "goflow/k8s/client"
	"goflow/k8s/pod/event/holder"
	"goflow/k8s/pod/inform"
	"goflow/k8s/pod/utils"
	"goflow/k8s/serviceaccount"

	"k8s.io/client-go/kubernetes"
)

// Orchestrator holds information for all DAGs
type Orchestrator struct {
	dagMapLock    *sync.RWMutex
	dagMap        map[string]*dagtype.DAG
	kubeClient    kubernetes.Interface
	config        *config.GoFlowConfig
	channelHolder *holder.ChannelHolder
}

func newOrchestratorFromClientAndConfig(
	client kubernetes.Interface,
	config *config.GoFlowConfig,
) *Orchestrator {
	return &Orchestrator{
		&sync.RWMutex{},
		make(map[string]*dagtype.DAG),
		client,
		config,
		holder.New(),
	}
}

// NewOrchestrator creates an empty instance of Orchestrator
func NewOrchestrator(configPath string) *Orchestrator {
	return newOrchestratorFromClientAndConfig(
		k8sclient.CreateKubeClient(),
		config.CreateConfig(configPath),
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
	orchestrator.dagMapLock.Lock()
	orchestrator.dagMap[dag.Config.Name] = dag
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
func (orchestrator Orchestrator) DAGs() []*dagtype.DAG {
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

// CollectDAGs fills up the dag map with existing dags
func (orchestrator *Orchestrator) CollectDAGs() {
	dagSlice := dagtype.GetDAGSFromFolder(
		orchestrator.config.DAGPath,
		orchestrator.kubeClient,
		*orchestrator.config,
	)
	for _, dag := range dagSlice {
		dagPresent := orchestrator.isDagPresent(*dag)
		if !dagPresent {
			orchestrator.addDAGServiceAccount(dag)
			orchestrator.AddDAG(dag)
		} else if dagPresent && orchestrator.isStoredDagDifferent(*dag) {
			logs.InfoLogger.Printf("Updating DAG %s which will run in namespace %s", dag.Config.Name, dag.Config.Namespace)
			logs.InfoLogger.Printf("Old DAG code: %s\n", orchestrator.GetDag(dag.Config.Name).Code)
			logs.InfoLogger.Printf("New DAG code: %s\n", dag.Code)
			// orchestrator.UpdateDag(&dag)
		}
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

// Start begins the orchestrator event loop
func (orchestrator *Orchestrator) Start(cycleDuration time.Duration, closingChannel chan struct{}) {
	taskInformer := orchestrator.getTaskInformer()
	taskInformer.Start()
	go cycleUntilChannelClose(
		orchestrator.CollectDAGs,
		closingChannel,
		cycleDuration,
		"Collect DAGs",
	)
	go cycleUntilChannelClose(orchestrator.RunDags, closingChannel, cycleDuration, "Run DAGs")
}
