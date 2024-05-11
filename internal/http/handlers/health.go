package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"g-fe-server/api/adapter/health"
	"g-fe-server/internal/http/middleware"
)

func HealthHandlers(nonFunctionalRouter *mux.Router, context context.Context) {

	ctxRoot := context.Value(CTX_CONTEXT_ROOT_KEY).(string)

	healthRouter := nonFunctionalRouter.PathPrefix("/health").Subrouter()
	healthRouter.Use(middleware.JSONResponse)

	healthRouter.Methods(http.MethodGet).HandlerFunc(onHealth).Path("").Name(ctxRoot + "/g/health")
	healthRouter.Methods(http.MethodGet).HandlerFunc(onHealth).Path("/")
}

func onHealth(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&health.HealthResponse{
		Status: health.Active,
	})
}
