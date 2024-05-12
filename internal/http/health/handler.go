package health

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	app_context "github.com/morphy76/g-fe-server/internal/http/context"
	"github.com/morphy76/g-fe-server/internal/http/middleware"
	"github.com/morphy76/g-fe-server/pkg/example"
)

func HealthHandlers(nonFunctionalRouter *mux.Router, context context.Context) {

	var (
		repository = context.Value(app_context.CTX_REPOSITORY_KEY).(example.Repository)
		ctxRoot    = context.Value(app_context.CTX_CONTEXT_ROOT_KEY).(app_context.ContextModel).ContextRoot
	)

	healthRouter := nonFunctionalRouter.Path("/health").Subrouter()
	healthRouter.Use(middleware.JSONResponse)

	healthRouter.Methods(http.MethodGet).HandlerFunc(onHealth(repository)).Name(ctxRoot + "/g/health")
}

func onHealth(repository example.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		overallStatus := Active

		repositoryStatus := Active
		repositoryCondition := repository.IsConnected() && repository.Ping()
		if !repositoryCondition {
			repositoryStatus = Inactive
			overallStatus = Inactive
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&HealthResponse{
			Status: overallStatus,
			SubSystems: map[string]HealthResponse{
				"Repository": {Status: repositoryStatus},
			},
		})
	}
}
