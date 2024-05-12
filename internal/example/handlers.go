package example

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	app_context "g-fe-server/internal/http/context"
	"g-fe-server/internal/http/middleware"
	"g-fe-server/pkg/example"
)

const (
	pathParamExampleId = "exampleId"
)

var repository = NewMemoryRepository()

func ExampleHandlers(apiRouter *mux.Router, context context.Context) {

	var (
		ctxRoot              = context.Value(app_context.CTX_CONTEXT_ROOT_KEY).(app_context.ContextModel).ContextRoot
		apiRoot              = fmt.Sprintf("%s/api/example", ctxRoot)
		apiParamExampleId    = fmt.Sprintf("{%s}", pathParamExampleId)
		apiResourceExampleId = fmt.Sprintf("%s/%s", apiRoot, apiParamExampleId)

		itemRouter = apiRouter.PathPrefix("/example").Subrouter()
	)

	itemRouter.Methods(http.MethodGet).HandlerFunc(onList).Path("").Name("GET " + apiRoot)
	itemRouter.Methods(http.MethodPost).HandlerFunc(onCreate).Name("POST " + apiRoot)
	itemRouter.Methods(http.MethodGet).HandlerFunc(onGet).Path("/" + apiParamExampleId).Name("GET " + apiResourceExampleId)
	itemRouter.Methods(http.MethodDelete).HandlerFunc(onDelete).Path("/" + apiParamExampleId).Name("DELETE " + apiResourceExampleId)
	itemRouter.Methods(http.MethodPut).HandlerFunc(onPut).Path("/" + apiParamExampleId).Name("PUT " + apiResourceExampleId)
}

func onList(w http.ResponseWriter, r *http.Request) {

	useLog := middleware.ExtractLoggerFromRequest(r, "example")
	useLog.Debug().Msg("Start listing examples")

	examples, err := repository.FindAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(examples)
	useLog.Info().Msg("End listing examples")
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
	exampleId := vars[pathParamExampleId]

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
	exampleId := vars[pathParamExampleId]

	err := repository.Delete(exampleId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func onPut(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	exampleId := vars[pathParamExampleId]

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
