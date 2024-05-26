package handlers

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/morphy76/g-fe-server/internal/example"
	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/morphy76/g-fe-server/internal/http/handlers/health"
	"github.com/morphy76/g-fe-server/internal/http/handlers/static"
	"github.com/morphy76/g-fe-server/internal/http/middleware"
	"github.com/morphy76/g-fe-server/internal/options"
)

func Handler(parent *mux.Router, context context.Context) {

	ctxRoot := context.Value(app_http.CTX_CONTEXT_SERVE_KEY).(*options.ServeOptions).ContextRoot

	// Parent router
	parent.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Trace().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Msg("Request received, context initialized")

			useRequest := r.WithContext(context)
			next.ServeHTTP(w, useRequest)
		})
	})

	// Context root router
	contextRouter := parent.PathPrefix(ctxRoot).Subrouter()
	contextRouter.Path("/ui").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, ctxRoot+"/ui/", http.StatusMovedPermanently)
	})
	if log.Trace().Enabled() {
		log.Trace().Msg("Context router registered")
	}

	// Static content
	staticRouter := contextRouter.PathPrefix("/ui/").Subrouter()
	if log.Trace().Enabled() {
		log.Trace().Msg("Static router registered")
	}
	staticRouter.Use(middleware.InjectSession)
	if log.Trace().Enabled() {
		log.Trace().Msg("Static middleware registered")
	}
	static.HandleStatic(staticRouter, context)
	if log.Trace().Enabled() {
		log.Trace().Msg("Static handler registered")
	}

	// API router
	apiRouter := contextRouter.PathPrefix("/api").Subrouter()
	if log.Trace().Enabled() {
		log.Trace().Msg("API router registered")
	}
	apiRouter.Use(middleware.TenantResolver)
	apiRouter.Use(middleware.RequestLogger)
	apiRouter.Use(middleware.JSONResponse)
	apiRouter.Use(mux.CORSMethodMiddleware(apiRouter))
	if log.Trace().Enabled() {
		log.Trace().Msg("API middleware registered")
	}

	// Non functional router
	nonFunctionalRouter := contextRouter.PathPrefix("/g").Subrouter()
	if log.Trace().Enabled() {
		log.Trace().Msg("Non functional router registered")
	}
	health.HealthHandlers(nonFunctionalRouter, context)
	if log.Trace().Enabled() {
		log.Trace().Msg("Health handler registered")
	}

	// Domain functions
	example.ExampleHandlers(apiRouter, context)
	if log.Trace().Enabled() {
		log.Trace().Msg("Example handler registered")
	}
}
