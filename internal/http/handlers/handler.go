package handlers

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"

	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/morphy76/g-fe-server/internal/http/handlers/auth"
	"github.com/morphy76/g-fe-server/internal/http/handlers/health"
	"github.com/morphy76/g-fe-server/internal/http/handlers/metrics"
	"github.com/morphy76/g-fe-server/internal/http/handlers/static"
	"github.com/morphy76/g-fe-server/internal/http/middleware"

	example_handlers "github.com/morphy76/g-fe-server/internal/example/http"
)

func Handler(parent *mux.Router, app_context context.Context) {

	serveOptions := app_http.ExtractServeOptions(app_context)
	dbOptions := app_http.ExtractDbOptions(app_context)
	sessionStore := app_http.ExtractSessionStore(app_context)
	dbClient := app_http.ExtractDb(app_context)
	relyingParty := app_http.ExtractRelyingParty(app_context)

	// Parent router
	parent.Use(func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			useRequest := r.WithContext(app_http.InjectDb(r.Context(), dbClient))
			useRequest = useRequest.WithContext(app_http.InjectDbOptions(useRequest.Context(), dbOptions))
			useRequest = useRequest.WithContext(app_http.InjectSessionStore(useRequest.Context(), sessionStore))
			useRequest = useRequest.WithContext(app_http.InjectServeOptions(useRequest.Context(), serveOptions))
			useRequest = useRequest.WithContext(app_http.InjectRelyingParty(useRequest.Context(), relyingParty))

			next.ServeHTTP(w, useRequest)
		})
	})

	// Non functional router
	nonFunctionalRouter := parent.PathPrefix("/g").Subrouter()
	if log.Trace().Enabled() {
		log.Trace().Msg("Non functional router registered")
	}
	health.HealthHandlers(nonFunctionalRouter, serveOptions.ContextRoot, dbOptions)
	if log.Trace().Enabled() {
		log.Trace().Msg("Health handler registered")
	}
	metrics.PrometheusHandlers(nonFunctionalRouter, serveOptions.ContextRoot)
	if log.Trace().Enabled() {
		log.Trace().Msg("Metrics handler registered")
	}

	// Context root router
	contextRouter := parent.PathPrefix(serveOptions.ContextRoot).Subrouter()
	contextRouter.Use(otelmux.Middleware("context",
		otelmux.WithPublicEndpoint(),
		otelmux.WithPropagators(otel.GetTextMapPropagator()),
	))
	contextRouter.Use(middleware.TenantResolver)
	contextRouter.Use(middleware.RequestLogger)
	contextRouter.Path("/ui").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, serveOptions.ContextRoot+"/ui/", http.StatusMovedPermanently)
	})
	if log.Trace().Enabled() {
		log.Trace().Msg("Context router registered")
	}

	// Auth router
	authRouter := contextRouter.PathPrefix("/auth").Subrouter()
	authRouter.Use(middleware.InjectSession)
	if log.Trace().Enabled() {
		log.Trace().Msg("Auth router registered")
	}
	auth.IAMHandlers(authRouter, serveOptions.ContextRoot, relyingParty)
	if log.Trace().Enabled() {
		log.Trace().Msg("Auth handler registered")
	}

	// Static content
	staticRouter := contextRouter.PathPrefix("/ui/").Subrouter()
	staticRouter.Use(middleware.InjectSession)
	staticRouter.Use(middleware.AuthenticationRequired)
	if log.Trace().Enabled() {
		log.Trace().Msg("Static router registered")
	}
	static.HandleStatic(staticRouter, serveOptions.ContextRoot, serveOptions.StaticPath)
	if log.Trace().Enabled() {
		log.Trace().Msg("Static handler registered")
	}

	// API router
	apiRouter := contextRouter.PathPrefix("/api").Subrouter()
	apiRouter.Use(mux.CORSMethodMiddleware(apiRouter))
	apiRouter.Use(middleware.JSONResponse)
	apiRouter.Use(middleware.PrometheusMiddleware)
	if log.Trace().Enabled() {
		log.Trace().Msg("API router registered")
	}

	// Domain functions
	example_handlers.ExampleHandlers(apiRouter, serveOptions.ContextRoot, dbOptions)
	if log.Trace().Enabled() {
		log.Trace().Msg("Example handler registered")
	}

	contextRouter.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		if len(route.GetName()) > 0 {
			router.Use(otelmux.Middleware(route.GetName(),
				otelmux.WithPublicEndpoint(),
				otelmux.WithPropagators(otel.GetTextMapPropagator()),
			))
		}
		return nil
	})
}
