package example

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	app_context "g-fe-server/internal/http/context"
	"g-fe-server/pkg/example"
)

var repository example.Repository = &MemoryRepository{}

func ExampleHandlers(apiRouter *mux.Router, context context.Context) {

	ctxRoot := context.Value(app_context.CTX_CONTEXT_ROOT_KEY).(app_context.ContextModel).ContextRoot

	itemRouter := apiRouter.PathPrefix("/example").Subrouter()

	itemRouter.Methods(http.MethodGet).HandlerFunc(OnList).Name(ctxRoot + "/api/example")
}

func OnList(w http.ResponseWriter, r *http.Request) {
	examples, err := repository.FindAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(examples)
}
