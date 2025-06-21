package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"

	"github.com/morphy76/g-fe-server/internal/auth"
	"github.com/morphy76/g-fe-server/internal/business/example"
	"github.com/morphy76/g-fe-server/internal/logger"
	"github.com/morphy76/g-fe-server/internal/server"
)

// Handler registers all HTTP handlers for the application
func Handler(
	appContext context.Context,
	rootRouter *mux.Router,
) error {
	routerLog := logger.GetLogger(appContext, "router")
	feServer, err := server.ExtractFEServer(appContext)
	if err != nil {
		return fmt.Errorf("failed to extract FEServer from context: %w", err)
	}

	// Parent router
	rootRouter.Use(otelmux.Middleware(feServer.ServiceName))

	err = initializeTheNonFunctionalRouter(appContext, rootRouter, feServer, routerLog)
	if err != nil {
		return fmt.Errorf("failed to initialize non-functional router: %w", err)
	}
	err = initializeTheFunctionalRouter(appContext, rootRouter, feServer, routerLog)
	if err != nil {
		return fmt.Errorf("failed to initialize functional router: %w", err)
	}

	return nil
}

func initializeTheNonFunctionalRouter(
	appContext context.Context,
	rootRouter *mux.Router,
	feServer *server.FEServer,
	routerLog zerolog.Logger,
) error {
	// propagates FEServer and logger to non functional requests
	// add non functional endopints
	// - health checks

	nfRoot := feServer.HTTPOpts.ServeOptions.NonFunctionalRoot

	nonFunctionalRouter := rootRouter.PathPrefix(nfRoot).Subrouter()
	err := enrichNonFunctionalRequestContext(appContext, nonFunctionalRouter)
	if err != nil {
		return fmt.Errorf("failed to enrich non-functional request context: %w", err)
	}
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Non functional router registered")
	}

	err = HandleHealth(appContext, nonFunctionalRouter, nfRoot, feServer.HealthChecksFn)
	if err != nil {
		return fmt.Errorf("failed to register health handler: %w", err)
	}

	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Health handler registered")
	}

	return nil
}

func enrichNonFunctionalRequestContext(appContext context.Context, router *mux.Router) error {

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			useRequestContext, err := server.InjectFEServer(r.Context(), appContext)
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to inject FEServer: %v", err), http.StatusInternalServerError)
				return
			}
			useRequestContext = logger.InjectLogger(useRequestContext, appContext)
			useRequest := r.WithContext(useRequestContext)
			next.ServeHTTP(w, useRequest)
		})
	})

	return nil
}

func initializeTheFunctionalRouter(
	appContext context.Context,
	rootRouter *mux.Router,
	feServer *server.FEServer,
	routerLog zerolog.Logger,
) error {
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

	ctxRoot := feServer.HTTPOpts.ServeOptions.ContextRoot

	contextRouter := rootRouter.PathPrefix(ctxRoot).Subrouter()
	err := enrichFunctionalRequestContext(appContext, contextRouter)
	if err != nil {
		return fmt.Errorf("failed to enrich functional request context: %w", err)
	}
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Context router registered")
	}

	err = HandleOpenAPI(contextRouter, ctxRoot)
	if err != nil {
		return fmt.Errorf("failed to register OpenAPI handler: %w", err)
	}
	err = addAuthHandlers(contextRouter, routerLog, feServer)
	if err != nil {
		return fmt.Errorf("failed to register auth handlers: %w", err)
	}
	err = addUIHandlers(contextRouter, feServer, routerLog)
	if err != nil {
		return fmt.Errorf("failed to register UI handlers: %w", err)
	}
	err = addAPIHandlers(contextRouter, feServer, routerLog)
	if err != nil {
		return fmt.Errorf("failed to register API handlers: %w", err)
	}

	return nil
}

func enrichFunctionalRequestContext(
	appContext context.Context,
	router *mux.Router,
) error {

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			useRequestContext, err := server.InjectFEServer(r.Context(), appContext)
			if err != nil {
				http.Error(w, fmt.Sprintf("failed to inject FEServer: %v", err), http.StatusInternalServerError)
				return
			}
			useRequestContext = logger.InjectLogger(useRequestContext, appContext)
			useRequest := r.WithContext(useRequestContext)
			next.ServeHTTP(w, useRequest)
		})
	})

	router.Use(logger.RequestLogger)
	return nil
}

func addAuthHandlers(
	contextRouter *mux.Router,
	routerLog zerolog.Logger,
	feServer *server.FEServer,
) error {
	// OIDC integration

	authRouter := contextRouter.PathPrefix("/auth").Subrouter()
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Auth router registered")
	}
	err := IAMHandlers(authRouter, feServer.HTTPOpts.ServeOptions, feServer.RelayingParty)
	if err != nil {
		return fmt.Errorf("failed to register IAM handlers: %w", err)
	}
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Auth handler registered")
	}
	return nil
}

func addUIHandlers(
	contextRouter *mux.Router,
	feServer *server.FEServer,
	routerLog zerolog.Logger,
) error {
	// Add UI endpoints for
	// - static content of the container application
	// - TODO: static content of MFEs

	staticRouter := contextRouter.PathPrefix("/ui").Subrouter()
	staticRouter.Use(auth.IsAuthenticated(
		feServer.SessionStore,
		feServer.HTTPOpts.SessionOptions.Name,
		feServer.RelayingParty,
		feServer.ResourceServer,
	))

	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Static router registered")
	}
	err := HandleStatic(staticRouter, feServer.HTTPOpts.ServeOptions.ContextRoot, feServer.HTTPOpts.ServeOptions.StaticPath)
	if err != nil {
		return fmt.Errorf("failed to register static handler: %w", err)
	}
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Static handler registered")
	}
	return nil
}

func addAPIHandlers(
	contextRouter *mux.Router,
	feServer *server.FEServer,
	routerLog zerolog.Logger,
) error {
	// Add API endpoints
	// - Resource modules bindings

	apiRouter := contextRouter.PathPrefix("/api").Subrouter()
	apiRouter.Use(auth.IsAuthenticated(
		feServer.SessionStore,
		feServer.HTTPOpts.SessionOptions.Name,
		feServer.RelayingParty,
		feServer.ResourceServer,
	))
	apiRouter.Use(setJSONResponse)

	err := bindModules(apiRouter, feServer, routerLog)
	if err != nil {
		return fmt.Errorf("failed to bind modules: %w", err)
	}
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("API router registered")
	}
	return nil
}

func bindModules(
	apiRouter *mux.Router,
	feServer *server.FEServer,
	routerLog zerolog.Logger,
) error {
	// Each resource module provides its own handlers
	// - example module

	example.Handler(apiRouter, feServer, routerLog)
	return nil
}
