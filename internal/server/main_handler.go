package server

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"

	"github.com/morphy76/g-fe-server/internal/http/handlers"
	"github.com/morphy76/g-fe-server/internal/http/middleware"
	"github.com/morphy76/g-fe-server/internal/http/session"
	"github.com/morphy76/g-fe-server/internal/logger"
)

// Handler registers all HTTP handlers for the application
func Handler(
	appContext context.Context,
	rootRouter *mux.Router,
) {
	routerLog := logger.GetLogger(appContext, "router")
	feServer := ExtractFEServer(appContext)

	// Parent router
	rootRouter.Use(otelmux.Middleware(feServer.ServiceName,
		otelmux.WithPublicEndpoint(),
	))

	initializeTheNonFunctionalRouter(appContext, rootRouter, feServer, routerLog)
	initializeTheFunctionalRouter(appContext, rootRouter, feServer, routerLog)
}

func initializeTheFunctionalRouter(appContext context.Context, rootRouter *mux.Router, feServer *FEServer, routerLog zerolog.Logger) {
	// Add functional endpoints
	// - static content (the UI) at /ui
	// - API endpoints at /api

	contextRouter := rootRouter.PathPrefix(feServer.ServeOpts.ContextRoot).Subrouter()
	enrichFunctionalRequestContext(contextRouter, feServer, appContext)
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Context router registered")
	}

	// TODO CORS: in the context router to allow MFE and APIs
	// contextRouter.Use(mux.CORSMethodMiddleware(apiRouter))
	// contextRouter.Use(middleware.TenantResolver)

	addAuthHandlers(contextRouter, routerLog, feServer)
	addUIHandlers(contextRouter, feServer, routerLog)
	addAPIHandlers(contextRouter, routerLog)
}

func addAuthHandlers(contextRouter *mux.Router, routerLog zerolog.Logger, feServer *FEServer) {
	authRouter := contextRouter.PathPrefix("/auth").Subrouter()
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Auth router registered")
	}
	handlers.IAMHandlers(authRouter, feServer.ServeOpts, feServer.RelayingParty)
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Auth handler registered")
	}
}

func addAPIHandlers(contextRouter *mux.Router, routerLog zerolog.Logger) {
	// serve server APIs

	apiRouter := contextRouter.PathPrefix("/api").Subrouter()
	apiRouter.Use(middleware.JSONResponse)

	// test API
	apiRouter.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		// <-time.After(1 * time.Second)
		// _, span := trace.SpanFromContext(r.Context()).TracerProvider().Tracer("mboh").Start(r.Context(), "testSpan")
		// defer span.End()
		// span.AddEvent("testEvent")
		w.Write([]byte("{\"message\": \"Hello, World!\"}"))
		// <-time.After(1 * time.Second)
		// span.RecordError(errors.New("testError"))
		// span.SetStatus(codes.Error, "testError")
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

func addUIHandlers(contextRouter *mux.Router, feServer *FEServer, routerLog zerolog.Logger) {
	// Static content
	staticRouter := contextRouter.PathPrefix("/ui").Subrouter()

	// staticRouter.Use(middleware.InjectSession(feServer.SessionStore, feServer.SessionsOpts.SessionName))
	// staticRouter.Use(middleware.HTTPSessionAuthenticationRequired(feServer.ServeOpts))
	// staticRouter.Use(middleware.HTTPSessionInspectAndRenew(feServer.ResourceServer, feServer.RelayingParty, feServer.ServeOpts))
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Static router registered")
	}
	handlers.HandleStatic(staticRouter, feServer.ServeOpts.ContextRoot, feServer.ServeOpts.StaticPath)
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Static handler registered")
	}
}

func initializeTheNonFunctionalRouter(appContext context.Context, rootRouter *mux.Router, feServer *FEServer, routerLog zerolog.Logger) {
	// add non functional endopints
	// - health checks

	nonFunctionalRouter := rootRouter.PathPrefix(feServer.ServeOpts.NonFunctionalRoot).Subrouter()
	enrichNonFunctionalRequestContext(nonFunctionalRouter, appContext)
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Non functional router registered")
	}
	// health checks to provide liveness and readiness endpoints
	handlers.HandleHealth(appContext, nonFunctionalRouter, feServer.ServeOpts.NonFunctionalRoot, feServer.HealthChecksFn)

	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Health handler registered")
	}
}

func enrichFunctionalRequestContext(router *mux.Router, feServer *FEServer, appContext context.Context) {

	router.Use(session.BindHTTPSessionToRequests(feServer.SessionStore, feServer.SessionName))

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			useRequestContext := InjectFEServer(r.Context(), appContext)
			useRequestContext = logger.InjectLogger(useRequestContext, appContext)
			useRequest := r.WithContext(useRequestContext)
			next.ServeHTTP(w, useRequest)
		})
	})

	router.Use(logger.RequestLogger)
}

func enrichNonFunctionalRequestContext(router *mux.Router, appContext context.Context) {

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			useRequestContext := InjectFEServer(r.Context(), appContext)
			useRequestContext = logger.InjectLogger(useRequestContext, appContext)
			useRequest := r.WithContext(useRequestContext)
			next.ServeHTTP(w, useRequest)
		})
	})
}
