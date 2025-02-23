package server

import (
	"context"
	"net/http"

	"github.com/google/uuid"
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

	enrichRequestContext(rootRouter, appContext)
	initializeTheNonFunctionalRouter(appContext, rootRouter, feServer, routerLog)
	initializeTheFunctionalRouter(rootRouter, feServer, routerLog)
}

func initializeTheFunctionalRouter(rootRouter *mux.Router, feServer *FEServer, routerLog zerolog.Logger) {
	// Add functional endpoints
	// - static content (the UI) at /ui
	// - API endpoints at /api

	contextRouter := rootRouter.PathPrefix(feServer.ServeOpts.ContextRoot).Subrouter()
	contextRouter.Use(session.BindHTTPSessionToRequests(feServer.SessionStore, feServer.SessionName))

	// Auth router
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

	// TODO CORS: in the context router to allow MFE and APIs
	// contextRouter.Use(mux.CORSMethodMiddleware(apiRouter))
	// contextRouter.Use(middleware.TenantResolver)
	addUIHandlers(contextRouter, feServer, routerLog)
	addAPIHandlers(contextRouter, routerLog)
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
	// serve static content
	contextRouter.Path("/ui").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, feServer.ServeOpts.ContextRoot+"/ui/", http.StatusTemporaryRedirect)
	})
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Context router registered")
	}

	// Static content
	staticRouter := contextRouter.PathPrefix("/ui/").Subrouter()
	staticRouter.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Fragment == "" {
				r.URL.Fragment = uuid.New().String()
				http.Redirect(w, r, r.URL.String(), http.StatusTemporaryRedirect)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	})
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

func enrichRequestContext(rootRouter *mux.Router, appContext context.Context) {
	// add the feServer and a logger to the request context

	rootRouter.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			useRequestContext := InjectFEServer(r.Context(), appContext)
			useRequestContext = logger.InjectLogger(useRequestContext, appContext)
			useRequest := r.WithContext(useRequestContext)
			next.ServeHTTP(w, useRequest)
		})
	})
	// middleware to trace HTTP requests and responses
	rootRouter.Use(logger.RequestLogger)
}
