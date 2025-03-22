package handlers

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"

	"github.com/morphy76/g-fe-server/internal/business/example"
	"github.com/morphy76/g-fe-server/internal/http/middleware"
	"github.com/morphy76/g-fe-server/internal/logger"
	"github.com/morphy76/g-fe-server/internal/server"
)

// Handler registers all HTTP handlers for the application
func Handler(
	appContext context.Context,
	rootRouter *mux.Router,
) {
	routerLog := logger.GetLogger(appContext, "router")
	feServer := server.ExtractFEServer(appContext)

	// Parent router
	rootRouter.Use(otelmux.Middleware(feServer.ServiceName))

	initializeTheNonFunctionalRouter(appContext, rootRouter, feServer, routerLog)
	initializeTheFunctionalRouter(appContext, rootRouter, feServer, routerLog)
}

func initializeTheNonFunctionalRouter(appContext context.Context, rootRouter *mux.Router, feServer *server.FEServer, routerLog zerolog.Logger) {
	// propagates FEServer and logger to non functional requests
	// add non functional endopints
	// - health checks

	nonFunctionalRouter := rootRouter.PathPrefix(feServer.ServeOpts.NonFunctionalRoot).Subrouter()
	enrichNonFunctionalRequestContext(nonFunctionalRouter, appContext)
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Non functional router registered")
	}

	HandleHealth(appContext, nonFunctionalRouter, feServer.ServeOpts.NonFunctionalRoot, feServer.HealthChecksFn)
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Health handler registered")
	}
}

func enrichNonFunctionalRequestContext(router *mux.Router, appContext context.Context) {

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			useRequestContext := server.InjectFEServer(r.Context(), appContext)
			useRequestContext = logger.InjectLogger(useRequestContext, appContext)
			useRequest := r.WithContext(useRequestContext)
			next.ServeHTTP(w, useRequest)
		})
	})
}

func initializeTheFunctionalRouter(appContext context.Context, rootRouter *mux.Router, feServer *server.FEServer, routerLog zerolog.Logger) {
	// propagates FEServer and logger to functional requests
	// Add functional endpoints
	// - OpenAPI as a public endpoint
	// - auth endpoints
	// - static content (the UI) at /ui
	// - API endpoints at /api
	// - TODO: HTTP session management for auth, UI, and API
	// - TODO: auth middleware (check bearer, fallback to HTTP session, inspect and renew)
	// - TODO: CORS (for API and UI X-Frame-Options)
	// - TODO: RBAC
	// - TODO: tenant resolution

	contextRouter := rootRouter.PathPrefix(feServer.ServeOpts.ContextRoot).Subrouter()
	enrichFunctionalRequestContext(contextRouter, feServer, appContext)
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Context router registered")
	}

	HandleOpenAPI(contextRouter, feServer.ServeOpts.ContextRoot)
	addAuthHandlers(contextRouter, routerLog, feServer)
	addUIHandlers(contextRouter, feServer, routerLog)
	addAPIHandlers(contextRouter, feServer, routerLog)
}

func enrichFunctionalRequestContext(router *mux.Router, feServer *server.FEServer, appContext context.Context) {

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			useRequestContext := server.InjectFEServer(r.Context(), appContext)
			useRequestContext = logger.InjectLogger(useRequestContext, appContext)
			useRequest := r.WithContext(useRequestContext)
			next.ServeHTTP(w, useRequest)
		})
	})

	router.Use(logger.RequestLogger)
}

func addAuthHandlers(contextRouter *mux.Router, routerLog zerolog.Logger, feServer *server.FEServer) {
	// OIDC integration

	authRouter := contextRouter.PathPrefix("/auth").Subrouter()
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Auth router registered")
	}
	IAMHandlers(authRouter, feServer.ServeOpts, feServer.RelayingParty)
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Auth handler registered")
	}
}

func addUIHandlers(contextRouter *mux.Router, feServer *server.FEServer, routerLog zerolog.Logger) {
	// Add UI endpoints for
	// - static content of the container application
	// - TODO: static content of MFEs

	staticRouter := contextRouter.PathPrefix("/ui").Subrouter()
	staticRouter.Use(middleware.IsAuthenticated(feServer))

	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Static router registered")
	}
	HandleStatic(staticRouter, feServer.ServeOpts.ContextRoot, feServer.ServeOpts.StaticPath)
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Static handler registered")
	}
}

func addAPIHandlers(contextRouter *mux.Router, feServer *server.FEServer, routerLog zerolog.Logger) {
	// Add API endpoints
	// - Resource modules bindings

	apiRouter := contextRouter.PathPrefix("/api").Subrouter()
	apiRouter.Use(middleware.IsAuthenticated(feServer))
	apiRouter.Use(middleware.JSONResponse)

	bindModules(apiRouter, feServer, routerLog)
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("API router registered")
	}
}

func bindModules(apiRouter *mux.Router, feServer *server.FEServer, routerLog zerolog.Logger) {
	// Each resource module provides its own handlers
	// - example module

	example.Handler(apiRouter, feServer, routerLog)
}
