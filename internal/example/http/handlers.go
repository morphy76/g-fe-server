package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/morphy76/g-fe-server/internal/example/api"
	"github.com/morphy76/g-fe-server/internal/options"
	"github.com/morphy76/g-fe-server/pkg/example"
)

const (
	pathParamExampleId = "exampleId"
)

func ExampleHandlers(apiRouter *mux.Router, ctxRoot string, dbOptions *options.DbOptions) {

	var (
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

var onList = api.ContextualizedApi(onContextualizedList)
var onCreate = api.ContextualizedApi(onContextualizedCreate)
var onGet = api.ContextualizedApi(onContextualizedGet)
var onDelete = api.ContextualizedApi(onContextualizedDelete)
var onPut = api.ContextualizedApi(onContextualizedPut)

func onContextualizedList(
	useLog zerolog.Logger,
	repository example.Repository,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		useLog.Trace().Msg("Start listing examples")
		defer func() {
			useLog.Info().Msg("End listing examples")
		}()

		examples, err := repository.FindAll()
		if err != nil {

			span := trace.SpanFromContext(r.Context())
			span.SetStatus(codes.Error, "FindAll failed")
			span.RecordError(err)

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

func onContextualizedCreate(
	useLog zerolog.Logger,
	repository example.Repository,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		useLog.Trace().Msg("Start creating example")
		defer func() {
			useLog.Info().Msg("End creating example")
		}()

		var e example.Example
		err := json.NewDecoder(r.Body).Decode(&e)
		if err != nil {

			span := trace.SpanFromContext(r.Context())
			span.SetStatus(codes.Error, "Create failed")
			span.RecordError(err)

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

			span := trace.SpanFromContext(r.Context())
			span.SetStatus(codes.Error, "Create failed")
			span.RecordError(err)

			useLog.Error().Msg(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func onContextualizedGet(
	useLog zerolog.Logger,
	repository example.Repository,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

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

			span := trace.SpanFromContext(r.Context())
			span.SetStatus(codes.Error, "Get failed")
			span.RecordError(err)

			useLog.Error().Msg(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(ex)
	}
}

func onContextualizedDelete(
	useLog zerolog.Logger,
	repository example.Repository,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

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

			span := trace.SpanFromContext(r.Context())
			span.SetStatus(codes.Error, "Delete failed")
			span.RecordError(err)

			useLog.Error().Msg(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func onContextualizedPut(
	useLog zerolog.Logger,
	repository example.Repository,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		useLog.Trace().Msg("Start updating example")
		defer func() {
			useLog.Info().Msg("End updating example")
		}()

		vars := mux.Vars(r)
		exampleId := vars[pathParamExampleId]

		var ex example.Example
		err := json.NewDecoder(r.Body).Decode(&ex)
		if err != nil {

			span := trace.SpanFromContext(r.Context())
			span.SetStatus(codes.Error, "Put failed")
			span.RecordError(err)

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

			span := trace.SpanFromContext(r.Context())
			span.SetStatus(codes.Error, "Put failed")
			span.RecordError(err)

			useLog.Error().Msg(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
