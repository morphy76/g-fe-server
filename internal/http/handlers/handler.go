package handlers

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"

	"github.com/morphy76/g-fe-server/internal/db"
	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/morphy76/g-fe-server/internal/http/handlers/health"
	"github.com/morphy76/g-fe-server/internal/http/handlers/static"
	"github.com/morphy76/g-fe-server/internal/http/middleware"
	"github.com/morphy76/g-fe-server/internal/options"
	"github.com/morphy76/g-fe-server/internal/serve"

	example_handlers "github.com/morphy76/g-fe-server/internal/example/http"
)

func Handler(parent *mux.Router, app_context context.Context) {

	serveOptions := app_context.Value(app_http.CTX_CONTEXT_SERVE_KEY).(*options.ServeOptions)
	dbOptions := app_context.Value(app_http.CTX_DB_OPTIONS_KEY).(*options.DbOptions)
	sessionStore := app_context.Value(app_http.CTX_SESSION_STORE_KEY).(sessions.Store)
	dbClient := app_context.Value(app_http.CTX_DB_KEY).(db.DbClient)

	// Parent router
	parent.Use(func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			useRequest := r.WithContext(context.WithValue(r.Context(), app_http.CTX_DB_KEY, dbClient))
			useRequest = useRequest.WithContext(context.WithValue(useRequest.Context(), app_http.CTX_DB_OPTIONS_KEY, dbOptions))
			useRequest = useRequest.WithContext(context.WithValue(useRequest.Context(), app_http.CTX_SESSION_STORE_KEY, sessionStore))
			useRequest = useRequest.WithContext(context.WithValue(useRequest.Context(), app_http.CTX_CONTEXT_SERVE_KEY, serveOptions))

			next.ServeHTTP(w, useRequest)
		})
	})
	parent.Use(middleware.TenantResolver)
	parent.Use(middleware.RequestLogger)

	// Context root router
	contextRouter := parent.PathPrefix(serveOptions.ContextRoot).Subrouter()
	contextRouter.Path("/ui").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, serveOptions.ContextRoot+"/ui/", http.StatusMovedPermanently)
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
	static.HandleStatic(staticRouter, serveOptions.ContextRoot, serveOptions.StaticPath)
	if log.Trace().Enabled() {
		log.Trace().Msg("Static handler registered")
	}

	// API router
	apiRouter := contextRouter.PathPrefix("/api").Subrouter()
	if log.Trace().Enabled() {
		log.Trace().Msg("API router registered")
	}

	apiRouter.Use(otelmux.Middleware(serve.OTEL_SERVICE_NAME,
		otelmux.WithPublicEndpoint(),
		otelmux.WithPropagators(otel.GetTextMapPropagator()),
	))
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
	health.HealthHandlers(nonFunctionalRouter, serveOptions.ContextRoot, dbOptions)
	if log.Trace().Enabled() {
		log.Trace().Msg("Health handler registered")
	}

	// Domain functions
	example_handlers.ExampleHandlers(apiRouter, serveOptions.ContextRoot, dbOptions)
	if log.Trace().Enabled() {
		log.Trace().Msg("Example handler registered")
	}
}
