package server

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"

	"github.com/morphy76/g-fe-server/internal/db"
	"github.com/morphy76/g-fe-server/internal/http/handlers/auth"
	"github.com/morphy76/g-fe-server/internal/http/handlers/health"
	"github.com/morphy76/g-fe-server/internal/http/handlers/static"
	"github.com/morphy76/g-fe-server/internal/http/middleware"
	"github.com/morphy76/g-fe-server/internal/logger"
	"github.com/morphy76/g-fe-server/internal/serve"
)

// Handler registers all HTTP handlers for the application
func Handler(
	appContext context.Context,
	rootRouter *mux.Router,
) {
	feServer := ExtractFEServer(appContext)
	routerLog := logger.GetLogger(appContext, "router")

	// Parent router
	rootRouter.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			useRequest := r.WithContext(appContext)
			next.ServeHTTP(w, useRequest)
		})
	})

	// Non functional router
	nonFunctionalRouter := rootRouter.PathPrefix(feServer.ServeOpts.NonFunctionalRoot).Subrouter()
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Non functional router registered")
	}
	// Add additional checks to test mongodb
	health.Handlers(appContext, nonFunctionalRouter, feServer.ServeOpts.NonFunctionalRoot,
		CreateHealthCheck(feServer.RelayingParty),
		db.CreateHealthCheck(feServer.DBOpts),
	)
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Health handler registered")
	}
	// metrics.PrometheusHandlers(nonFunctionalRouter, serveOptions.ContextRoot)
	// if log.Trace().Enabled() {
	// 	log.Trace().
	// 		Msg("Metrics handler registered")
	// }

	// Context root router with OTel
	contextRouter := rootRouter.PathPrefix(feServer.ServeOpts.ContextRoot).Subrouter()
	contextRouter.Use(otelmux.Middleware(serve.OTelAppName,
		otelmux.WithPublicEndpoint(),
		otelmux.WithPropagators(otel.GetTextMapPropagator()),
		otelmux.WithTracerProvider(otel.GetTracerProvider()),
	))
	// TODO CORS: in the context router to allow MFE and APIs
	// contextRouter.Use(mux.CORSMethodMiddleware(apiRouter))
	// contextRouter.Use(middleware.TenantResolver)
	contextRouter.Use(middleware.RequestLogger)
	contextRouter.Path("/ui").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, feServer.ServeOpts.ContextRoot+"/ui/", http.StatusTemporaryRedirect)
	})
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Context router registered")
	}

	// Auth router
	authRouter := contextRouter.PathPrefix("/auth").Subrouter()
	authRouter.Use(middleware.InjectSession(feServer.SessionStore, feServer.ServeOpts.SessionName))
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Auth router registered")
	}
	auth.IAMHandlers(authRouter, feServer.ServeOpts, feServer.RelayingParty)
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Auth handler registered")
	}

	// Static content
	staticRouter := contextRouter.PathPrefix("/ui/").Subrouter()
	staticRouter.Use(middleware.InjectSession(feServer.SessionStore, feServer.ServeOpts.SessionName))
	staticRouter.Use(middleware.HTTPSessionAuthenticationRequired(feServer.ServeOpts))
	staticRouter.Use(middleware.HTTPSessionInspectAndRenew(feServer.ResourceServer, feServer.RelayingParty, feServer.ServeOpts))
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Static router registered")
	}
	static.HandleStatic(staticRouter, feServer.ServeOpts.ContextRoot, feServer.ServeOpts.StaticPath)
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Static handler registered")
	}

	// API router
	apiRouter := contextRouter.PathPrefix("/api").Subrouter()
	apiRouter.Use(middleware.JSONResponse)
	apiRouter.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!\n"))
	})
	// apiRouter.Use(middleware.PrometheusMiddleware)
	// TODO: gw oriented auth, inspect and renew
	// apiRouter.Use(middleware.InjectSession(feServer.SessionStore, feServer.ServeOpts.SessionName)) ????
	// apiRouter.Use(middleware.MixedAuthenticationRequired)
	// apiRouter.Use(middleware.MixedInspectAndRenew)
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("API router registered")
	}
}
