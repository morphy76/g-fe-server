package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/morphy76/g-fe-server/internal/common/health"
	"github.com/morphy76/g-fe-server/internal/http/middleware"
	"github.com/morphy76/g-fe-server/internal/logger"
)

// Handlers registers the health check handlers
func Handlers(
	appContext context.Context,
	parent *mux.Router,
	ctxRoot string,
	additionalChecks []health.AdditionalCheckFn,
) {
	healthRouter := parent.PathPrefix("/health").Subrouter()
	healthRouter.Use(middleware.JSONResponse)

	liveChecks := make([]health.HealthCheckFn, 0)
	readyChecks := make([]health.HealthCheckFn, 0)

	for _, check := range additionalChecks {
		checkFn, probe := check(appContext)
		if probe&health.Live != 0 {
			liveChecks = append(liveChecks, checkFn)
		}
		if probe&health.Ready != 0 {
			readyChecks = append(readyChecks, checkFn)
		}
	}

	liveRouter := healthRouter.PathPrefix("/live").Subrouter()
	liveRouter.Methods(http.MethodGet).HandlerFunc(onHealth(liveChecks)).Name("GET " + ctxRoot + "/health/live")

	readyRouter := healthRouter.PathPrefix("/ready").Subrouter()
	readyRouter.Methods(http.MethodGet).HandlerFunc(onHealth(readyChecks)).Name("GET " + ctxRoot + "/health/ready")
}

func onHealth(additionalChecks []health.HealthCheckFn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		overallStatus := health.Active

		timeoutContext, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		log := logger.GetLogger(r.Context(), "health")

		var subsystems = make(map[string]health.HealthResponse)
		for _, check := range additionalChecks {
			label, status := check(timeoutContext)
			if status == health.Inactive {
				overallStatus = health.Inactive
			}
			subsystems[label] = health.HealthResponse{
				Status: status,
			}
		}

		if overallStatus == health.Active {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		healthResponse := &health.HealthResponse{
			Status:     overallStatus,
			SubSystems: subsystems,
		}

		if healthResponse.Status == health.Inactive {
			log.Warn().Interface("health", healthResponse).Msg("health check")
		} else {
			log.Trace().Interface("health", healthResponse).Msg("health check")
		}

		json.NewEncoder(w).Encode(healthResponse)
	}
}
