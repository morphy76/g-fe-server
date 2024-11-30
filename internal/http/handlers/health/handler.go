package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/morphy76/g-fe-server/internal/http/middleware"
	"github.com/morphy76/g-fe-server/internal/logger"
)

// Handlers registers the health check handlers
func Handlers(
	appContext context.Context,
	parent *mux.Router,
	ctxRoot string,
	additionalChecks ...app_http.AdditionalCheckFn,
) {
	healthRouter := parent.PathPrefix("/health").Subrouter()
	healthRouter.Use(middleware.JSONResponse)

	liveChecks := make([]app_http.HealthCheckFn, 0)
	readyChecks := make([]app_http.HealthCheckFn, 0)

	for _, check := range additionalChecks {
		checkFn, probe := check(appContext)
		if probe&app_http.Live != 0 {
			liveChecks = append(liveChecks, checkFn)
		}
		if probe&app_http.Ready != 0 {
			readyChecks = append(readyChecks, checkFn)
		}
	}

	liveRouter := healthRouter.PathPrefix("/live").Subrouter()
	liveRouter.Methods(http.MethodGet).HandlerFunc(onHealth(liveChecks)).Name("GET " + ctxRoot + "/health/live")

	readyRouter := healthRouter.PathPrefix("/ready").Subrouter()
	readyRouter.Methods(http.MethodGet).HandlerFunc(onHealth(readyChecks)).Name("GET " + ctxRoot + "/health/ready")
}

func onHealth(additionalChecks []app_http.HealthCheckFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		overallStatus := app_http.Active

		timeoutContext, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		log := logger.GetLogger(r.Context(), "health")

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

		healthResponse := &app_http.HealthResponse{
			Status:     overallStatus,
			SubSystems: subsystems,
		}

		if healthResponse.Status == app_http.Inactive {
			log.Warn().Interface("health", healthResponse).Msg("health check")
		}

		json.NewEncoder(w).Encode(healthResponse)
	}
}
