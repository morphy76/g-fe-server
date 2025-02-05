package server

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/morphy76/g-fe-server/internal/http/handlers/health"
	"github.com/morphy76/g-fe-server/internal/http/handlers/static"
	"github.com/morphy76/g-fe-server/internal/http/middleware"
	"github.com/morphy76/g-fe-server/internal/logger"
)

// Handler registers all HTTP handlers for the application
func Handler(
	appContext context.Context,
	rootRouter *mux.Router,
) {
	feServer := ExtractFEServer(appContext)
	routerLog := logger.GetLogger(appContext, "router")

	// rootRouter provides OTEL, application context facilities and the request logger; it splits into:
	// - Non functional router for health checks and metrics
	// - Context router for
	//  - Auth router for OIDC authentication, with HTTP session
	//  - Static router for serving static content, with HTTP session and authenticated based on HTTP session
	//  - API router for serving APIs, with default JSON response content

	// Parent router
	// enrich the context for OTEL tracing
	// rootRouter.Use(otelmux.Middleware(feServer.OTelOpts.ServiceName,
	// 	otelmux.WithPublicEndpoint(),
	// ))
	// enrich the request context with logger and server instance extracting from the application context
	rootRouter.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			useRequestContext := InjectFEServer(r.Context(), appContext)
			useRequestContext = logger.InjectLogger(useRequestContext, appContext)
			useRequest := r.WithContext(useRequestContext)
			next.ServeHTTP(w, useRequest)
		})
	})
	// middleware to trace HTTP requests and responses
	rootRouter.Use(middleware.RequestLogger)

	// Non functional router
	nonFunctionalRouter := rootRouter.PathPrefix(feServer.ServeOpts.NonFunctionalRoot).Subrouter()
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Non functional router registered")
	}
	// health checks to provide liveness and readiness endpoints
	health.Handlers(appContext, nonFunctionalRouter, feServer.ServeOpts.NonFunctionalRoot) // CreateHealthCheck(feServer.RelayingParty),
	// db.CreateHealthCheck(feServer.DBOpts),

	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Health handler registered")
	}
	// metrics.PrometheusHandlers(nonFunctionalRouter, serveOptions.ContextRoot)
	// if log.Trace().Enabled() {
	// 	log.Trace().
	// 		Msg("Metrics handler registered")
	// }

	// Context root router
	contextRouter := rootRouter.PathPrefix(feServer.ServeOpts.ContextRoot).Subrouter()
	// TODO CORS: in the context router to allow MFE and APIs
	// contextRouter.Use(mux.CORSMethodMiddleware(apiRouter))
	// contextRouter.Use(middleware.TenantResolver)
	contextRouter.Path("/ui").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, feServer.ServeOpts.ContextRoot+"/ui/", http.StatusTemporaryRedirect)
	})
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Context router registered")
	}

	// Auth router
	// authRouter := contextRouter.PathPrefix("/auth").Subrouter()
	// authRouter.Use(middleware.InjectSession(feServer.SessionStore, feServer.SessionsOpts.SessionName))
	// if routerLog.Trace().Enabled() {
	// 	routerLog.Trace().
	// 		Msg("Auth router registered")
	// }
	// auth.IAMHandlers(authRouter, feServer.ServeOpts, feServer.RelayingParty)
	// if routerLog.Trace().Enabled() {
	// 	routerLog.Trace().
	// 		Msg("Auth handler registered")
	// }

	// Static content
	staticRouter := contextRouter.PathPrefix("/ui/").Subrouter()
	// staticRouter.Use(middleware.InjectSession(feServer.SessionStore, feServer.SessionsOpts.SessionName))
	// staticRouter.Use(middleware.HTTPSessionAuthenticationRequired(feServer.ServeOpts))
	// staticRouter.Use(middleware.HTTPSessionInspectAndRenew(feServer.ResourceServer, feServer.RelayingParty, feServer.ServeOpts))
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
