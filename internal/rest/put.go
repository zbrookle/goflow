package rest

import (
	"encoding/json"
	"fmt"
	dagconfig "goflow/internal/dag/config"
	"goflow/internal/dag/orchestrator"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

func registerPutHandles(orch *orchestrator.Orchestrator, router *mux.Router) {
	router.HandleFunc("/dag", func(w http.ResponseWriter, r *http.Request) {
		dagConfig := &dagconfig.DAGConfig{}
		requestBytes := make([]byte, 0)
		requestBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		json.Unmarshal(requestBytes, dagConfig)
		fmt.Println(fmt.Sprint(dagConfig))
		orch.WriteDAGFile(dagConfig)
	}).Methods("POST")
}
