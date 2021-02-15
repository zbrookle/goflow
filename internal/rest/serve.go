package rest

import (
	"fmt"
	"goflow/internal/dag/orchestrator"
	"net/http"

	"github.com/gorilla/mux"
)

// Serve registers handlers and starts the goflow webserver
func Serve(host string, port int, orchestrator orchestrator.Orchestrator) {
	router := mux.NewRouter()
	registerGetHandles(orchestrator, router)
	http.Handle("/", router)
	http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil)
}
