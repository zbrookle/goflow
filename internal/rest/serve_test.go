package rest

import (
	"goflow/internal/dag/orchestrator"

	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// serveSingle is for testing only
func serveSingle(
	host string,
	port int,
	orchestrator *orchestrator.Orchestrator,
	handlerModifier func(*orchestrator.Orchestrator, *mux.Router),
) {
	router := mux.NewRouter()
	registerGetHandles(orchestrator, router)
	http.Handle("/", router)
	http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
}
