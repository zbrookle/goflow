package rest

import (
	"fmt"
	"goflow/internal/dag/orchestrator"

	"net/http"

	"github.com/gorilla/mux"
)

func registerPutHandles(orch *orchestrator.Orchestrator, router *mux.Router) {
	router.HandleFunc(
		"/dag/{name}/toggle",
		func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			name := vars["name"]
			dag := orch.GetDag(name)
			if dag == nil {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, "DAG not found!")
				return
			}
			dag.ToggleOnOff()
			fmt.Fprint(w, "DAG toggle success")
		},
	).Methods(
		http.MethodPut,
	)
}
