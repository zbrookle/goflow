package rest

import (
	"fmt"
	"goflow/internal/dag/orchestrator"
	"net/http"

	"github.com/gorilla/mux"
)

func setHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

// Serve registers handlers and starts the goflow webserver
func Serve(host string, port int, orchestrator *orchestrator.Orchestrator) {
	router := mux.NewRouter()
	router.Methods(http.MethodOptions).HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			setHeaders(w)
			w.Header().Set("Access-Control-Allow-Methods", http.MethodPut)
		},
	)
	registerGetHandles(orchestrator, router)
	registerPostHandles(orchestrator, router)
	registerPutHandles(orchestrator, router)
	http.Handle("/", router)
	http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), router)
}
