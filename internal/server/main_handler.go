package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

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
	rootRouter.Use(otelmux.Middleware(feServer.ServiceName))

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

	// test APIs & functions
	apiRouter.HandleFunc("/up", func(w http.ResponseWriter, r *http.Request) {
		// this API has its own span started by the OTEL SDK integration with the mux router
		useLogger := logger.GetLogger(r.Context(), "test")

		feServer := ExtractFEServer(r.Context())
		isTestFeatOn := feServer.IsFeatureEnabled("test")

		// this is an inner span, started by the application, representing upstream logic
		_, span := trace.SpanFromContext(r.Context()).TracerProvider().Tracer("mboh").Start(r.Context(), "upBiz")
		useLogger.Debug().Msg("start up biz")
		<-time.After(1 * time.Second)
		span.AddEvent("testEventUp")
		// session := session.ExtractSession(r.Context())
		// session.Put("test", uuid.New().String())
		useLogger.Info().Msg("end up biz")
		span.End()

		if isTestFeatOn {
			// this other span represents remote downstream logic
			_, span = trace.SpanFromContext(r.Context()).TracerProvider().Tracer("mboh").Start(r.Context(), "downBiz")
			useLogger.Debug().Msg("start down biz")
			<-time.After(1 * time.Second)

			// a generic request to the same server, but to a different endpoint
			newUrl := fmt.Sprintf("http://%s%s",
				r.Host,
				strings.Replace(r.URL.Path, "up", "down", 1),
			)
			newReq, err := http.NewRequestWithContext(r.Context(), http.MethodGet, newUrl, nil)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// mediated by the AIW facade to inject the OTEL context
			newRes, err := feServer.GetAIWFacade().Call(r.Context(), newReq)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// the request span, the HTTP call namely, is closed when closing the response body
			transferBodyTest(w, newRes)
			// the downstream span can be closed at the end
			defer span.End()
			<-time.After(1 * time.Second)
			span.AddEvent("testEventDown")
			useLogger.Info().Msg("end down biz")
		} else {
			w.Write([]byte("{\"message\": \"Hello, World!\"}"))
		}
	})
	apiRouter.HandleFunc("/down", func(w http.ResponseWriter, r *http.Request) {
		// this API has its own span started by the OTEL SDK integration with the mux router
		// but it seems not to be bound to the client parent span (TODO)
		useLogger := logger.GetLogger(r.Context(), "test")

		feServer := ExtractFEServer(r.Context())

		span := trace.SpanFromContext(r.Context())

		useLogger.Debug().Msg("start downer biz")
		// session := session.ExtractSession(r.Context())
		// test := session.Get("test").(string)
		test := uuid.NewString()
		// TODO mongo client should be bound to the OTEL context WIP
		feServer.MongoClient.Database("fe_db").Collection("test_collection").InsertOne(r.Context(), map[string]string{"test": test})
		span.AddEvent("testEventDowner")
		w.Write([]byte("{\"message\": \"Hello, World, " + test + "!\"}"))
		<-time.After(1 * time.Second)
		span.RecordError(errors.New("testError"))
		span.SetStatus(codes.Error, "testError")
		useLogger.Info().Msg("end downer biz")
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

// this method grants to close the http request span accordingly to the completion of managing the response body
// not really interested in handling error so far
func transferBodyTest(w http.ResponseWriter, newRes *http.Response) bool {
	w.WriteHeader(newRes.StatusCode)
	body, err := io.ReadAll(newRes.Body)
	// example of business metrics
	addBusinessMetrics(len(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return true
	}
	w.Write(body)
	defer newRes.Body.Close()
	return false
}

func addBusinessMetrics(tokens int) {
	meter := otel.GetMeterProvider().Meter("fe_server.metrics")

	var totalTokens int64
	tokensCounter, err := meter.Int64ObservableCounter(
		"fe_server.tokens",
		metric.WithDescription("The number of tokens processed"),
		metric.WithUnit("10"),
	)
	if err == nil {
		meter.RegisterCallback(
			func(ctx context.Context, observer metric.Observer) error {
				totalTokens += int64(tokens)
				observer.ObserveInt64(
					tokensCounter,
					totalTokens,
					metric.WithAttributes(
						attribute.Bool("billable", true),
						attribute.String("tenant", "todo"),
						attribute.String("subscription", "todo"),
					),
				)
				return nil
			},
			tokensCounter,
		)
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
