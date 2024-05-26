package example

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/morphy76/g-fe-server/internal/http/middleware"
	"github.com/morphy76/g-fe-server/internal/options"
	"github.com/morphy76/g-fe-server/pkg/example"
)

const (
	pathParamExampleId = "exampleId"
)

func ExampleHandlers(apiRouter *mux.Router, context context.Context) {

	var (
		repository           = context.Value(app_http.CTX_REPOSITORY_KEY).(example.Repository)
		ctxRoot              = context.Value(app_http.CTX_CONTEXT_ROOT_KEY).(options.ServeOptions).ContextRoot
		apiRoot              = fmt.Sprintf("%s/api/example", ctxRoot)
		apiParamExampleId    = fmt.Sprintf("{%s}", pathParamExampleId)
		apiResourceExampleId = fmt.Sprintf("%s/%s", apiRoot, apiParamExampleId)

		itemRouter = apiRouter.PathPrefix("/example").Subrouter()
	)

	itemRouter.Methods(http.MethodGet).HandlerFunc(onList(repository)).Path("").Name("GET " + apiRoot)
	itemRouter.Methods(http.MethodPost).HandlerFunc(onCreate(repository)).Name("POST " + apiRoot)
	itemRouter.Methods(http.MethodGet).HandlerFunc(onGet(repository)).Path("/" + apiParamExampleId).Name("GET " + apiResourceExampleId)
	itemRouter.Methods(http.MethodDelete).HandlerFunc(onDelete(repository)).Path("/" + apiParamExampleId).Name("DELETE " + apiResourceExampleId)
	itemRouter.Methods(http.MethodPut).HandlerFunc(onPut(repository)).Path("/" + apiParamExampleId).Name("PUT " + apiResourceExampleId)
}

func onList(repository example.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		useLog := middleware.ExtractLoggerFromRequest(r, "example")
		useLog.Trace().Msg("Start listing examples")
		defer func() {
			useLog.Info().Msg("End listing examples")
		}()

		examples, err := repository.FindAll()
		if err != nil {
			useLog.Error().Msg(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if len(examples) > 0 {
			json.NewEncoder(w).Encode(examples)
		} else {
			w.Write([]byte("[]"))
		}
	}
}

func onCreate(repository example.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		useLog := middleware.ExtractLoggerFromRequest(r, "example")
		useLog.Trace().Msg("Start creating example")
		defer func() {
			useLog.Info().Msg("End creating example")
		}()

		var e example.Example
		err := json.NewDecoder(r.Body).Decode(&e)
		if err != nil {
			useLog.Error().Msg(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = repository.Save(e)
		if err != nil {
			if example.IsAlreadyExists(err) {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}
			useLog.Error().Msg(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func onGet(repository example.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		useLog := middleware.ExtractLoggerFromRequest(r, "example")
		useLog.Trace().Msg("Start fetching example")
		defer func() {
			useLog.Info().Msg("End fetching example")
		}()

		vars := mux.Vars(r)
		exampleId := vars[pathParamExampleId]

		ex, err := repository.FindById(exampleId)
		if err != nil {
			if example.IsNotFound(err) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			useLog.Error().Msg(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ex)
	}
}

func onDelete(repository example.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		useLog := middleware.ExtractLoggerFromRequest(r, "example")
		useLog.Trace().Msg("Start deleting example")
		defer func() {
			useLog.Info().Msg("End deleting example")
		}()

		vars := mux.Vars(r)
		exampleId := vars[pathParamExampleId]

		err := repository.Delete(exampleId)
		if err != nil {
			if example.IsNotFound(err) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			useLog.Error().Msg(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func onPut(repository example.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		useLog := middleware.ExtractLoggerFromRequest(r, "example")
		useLog.Trace().Msg("Start updating example")
		defer func() {
			useLog.Info().Msg("End updating example")
		}()

		vars := mux.Vars(r)
		exampleId := vars[pathParamExampleId]

		var ex example.Example
		err := json.NewDecoder(r.Body).Decode(&ex)
		if err != nil {
			useLog.Error().Msg(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ex.Name = exampleId
		err = repository.Update(ex)
		if err != nil {
			if example.IsNotFound(err) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			useLog.Error().Msg(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
