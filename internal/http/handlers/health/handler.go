package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/morphy76/g-fe-server/internal/http/middleware"
	"github.com/morphy76/g-fe-server/internal/options"
)

// Handlers registers the health check handlers
func Handlers(
	appContext context.Context,
	parent *mux.Router,
	serveOptions *options.ServeOptions,
	additionalChecks ...app_http.HealthCheckFn,
) {
	ctxRoot := serveOptions.ContextRoot

	healthRouter := parent.Path("/health").Subrouter()
	healthRouter.Use(middleware.JSONResponse)

	healthRouter.Methods(http.MethodGet).HandlerFunc(onHealth(additionalChecks)).Name("GET " + ctxRoot + "/health")
}

func onHealth(additionalChecks []app_http.HealthCheckFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		overallStatus := app_http.Active

		timeoutContext, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		var subsystems = make(map[string]app_http.HealthResponse)
		for _, check := range additionalChecks {
			label, status := check(timeoutContext)
			if status == app_http.Inactive {
				overallStatus = app_http.Inactive
			}
			subsystems[label] = app_http.HealthResponse{
				Status: status,
			}
		}

		if overallStatus == app_http.Active {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		json.NewEncoder(w).Encode(&app_http.HealthResponse{
			Status:     overallStatus,
			SubSystems: subsystems,
		})
	}
}
