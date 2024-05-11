package health

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	app_context "g-fe-server/internal/http/context"
	"g-fe-server/internal/http/middleware"
)

func HealthHandlers(nonFunctionalRouter *mux.Router, context context.Context) {

	ctxRoot := context.Value(app_context.CTX_CONTEXT_ROOT_KEY).(app_context.ContextModel).ContextRoot

	healthRouter := nonFunctionalRouter.Path("/health").Subrouter()
	healthRouter.Use(middleware.JSONResponse)

	healthRouter.Methods(http.MethodGet).HandlerFunc(onHealth).Name(ctxRoot + "/g/health")
}

func onHealth(w http.ResponseWriter, r *http.Request) {

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&HealthResponse{
		Status: Active,
	})
}
