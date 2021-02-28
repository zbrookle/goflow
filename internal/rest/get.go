package rest

import (
	"fmt"
	"goflow/internal/dag/dagtype"
	"goflow/internal/dag/orchestrator"
	"net/http"
	"sort"

	"github.com/gorilla/mux"
)

const missingDagMsg = "\"There is no DAG with given name\""

func getDagFromRequest(
	orch *orchestrator.Orchestrator,
	w http.ResponseWriter,
	r *http.Request,
) *dagtype.DAG {
	vars := mux.Vars(r)
	dagName := vars["name"]
	dag := orch.GetDag(dagName)
	if dag == nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, missingDagMsg)
		return nil
	}
	setHeaders(w)
	return dag
}

func registerGetHandles(orch *orchestrator.Orchestrator, router *mux.Router) {

	router.HandleFunc("/dags", func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w)
		dags := orch.DAGs()
		sort.Sort(dagtype.ByName(dags))
		fmt.Fprint(w, dags)
	})

	router.HandleFunc("/dag/{name}", func(w http.ResponseWriter, r *http.Request) {
		dag := getDagFromRequest(orch, w, r)
		if dag == nil {
			return
		}
		fmt.Fprintf(w, dag.String())
	})

	router.HandleFunc("/dag/{name}/runs", func(w http.ResponseWriter, r *http.Request) {
		dag := getDagFromRequest(orch, w, r)
		if dag == nil {
			return
		}
		fmt.Fprint(w, dag.DAGRuns)
	})

	router.HandleFunc("/dag/{name}/metrics", func(w http.ResponseWriter, r *http.Request) {
		dag := getDagFromRequest(orch, w, r)
		if dag == nil {
			return
		}
		fmt.Fprint(w, dag.Metrics())
	})
}
