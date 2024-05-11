package example

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	app_context "g-fe-server/internal/http/context"
	"g-fe-server/pkg/example"
)

var repository = NewMemoryRepository()

func ExampleHandlers(apiRouter *mux.Router, context context.Context) {

	ctxRoot := context.Value(app_context.CTX_CONTEXT_ROOT_KEY).(app_context.ContextModel).ContextRoot

	itemRouter := apiRouter.PathPrefix("/example").Subrouter()

	itemRouter.Methods(http.MethodGet).HandlerFunc(onList).Path("").Name("GET" + ctxRoot + "/api/example")
	itemRouter.Methods(http.MethodPost).HandlerFunc(onCreate).Name("POST" + ctxRoot + "/api/example")
	itemRouter.Methods(http.MethodGet).HandlerFunc(onGet).Path("/{exampleId}").Name("GET" + ctxRoot + "/api/example/{exampleId}")
	itemRouter.Methods(http.MethodDelete).HandlerFunc(onDelete).Path("/{exampleId}").Name("DELETE" + ctxRoot + "/api/example/{exampleId}")
	itemRouter.Methods(http.MethodPut).HandlerFunc(onPut).Path("/{exampleId}").Name("PUT" + ctxRoot + "/api/example/{exampleId}")
}

func onList(w http.ResponseWriter, r *http.Request) {
	examples, err := repository.FindAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(examples)
}

func onCreate(w http.ResponseWriter, r *http.Request) {
	var example example.Example
	err := json.NewDecoder(r.Body).Decode(&example)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = repository.Save(example)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func onGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	exampleId := vars["exampleId"]

	example, err := repository.FindById(exampleId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(example)
}

func onDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	exampleId := vars["exampleId"]

	err := repository.Delete(exampleId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func onPut(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	exampleId := vars["exampleId"]

	var example example.Example
	err := json.NewDecoder(r.Body).Decode(&example)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	example.Name = exampleId
	err = repository.Update(example)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
