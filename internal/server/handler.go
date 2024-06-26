package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/client/rs"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/morphy76/g-fe-server/internal/http/handlers/auth"
	"github.com/morphy76/g-fe-server/internal/http/handlers/health"
	"github.com/morphy76/g-fe-server/internal/http/handlers/metrics"
	"github.com/morphy76/g-fe-server/internal/http/handlers/static"
	"github.com/morphy76/g-fe-server/internal/http/middleware"
	"github.com/morphy76/g-fe-server/internal/serve"
)

func Handler(
	parent *mux.Router,
	app_context context.Context,
) *mux.Router {
	serveOptions := app_http.ExtractServeOptions(app_context)
	sessionStore := app_http.ExtractSessionStore(app_context)
	oidcOptions := app_http.ExtractOidcOptions(app_context)

	var relyingParty rp.RelyingParty
	var resourceServer rs.ResourceServer
	if !oidcOptions.Disabled {
		relyingParty = app_http.ExtractRelyingParty(app_context)
		resourceServer = app_http.ExtractOidcResource(app_context)
	}

	// Parent router
	parent.Use(otelmux.Middleware(serve.OTEL_GW_NAME,
		otelmux.WithPublicEndpoint(),
		otelmux.WithPropagators(otel.GetTextMapPropagator()),
		otelmux.WithTracerProvider(otel.GetTracerProvider()),
	))
	parent.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			useRequest := r.WithContext(app_http.InjectSessionStore(r.Context(), sessionStore))
			useRequest = useRequest.WithContext(app_http.InjectServeOptions(useRequest.Context(), serveOptions))
			useRequest = useRequest.WithContext(app_http.InjectOidcOptions(useRequest.Context(), oidcOptions))
			if !oidcOptions.Disabled {
				useRequest = useRequest.WithContext(app_http.InjectRelyingParty(useRequest.Context(), relyingParty))
				useRequest = useRequest.WithContext(app_http.InjectOidcResource(useRequest.Context(), resourceServer))
			}

			next.ServeHTTP(w, useRequest)
		})
	})

	// Non functional router
	nonFunctionalRouter := parent.PathPrefix("/g").Subrouter()
	if log.Trace().Enabled() {
		log.Trace().
			Msg("Non functional router registered")
	}
	health.HealthHandlers(nonFunctionalRouter, app_context)
	if log.Trace().Enabled() {
		log.Trace().
			Msg("Health handler registered")
	}
	metrics.PrometheusHandlers(nonFunctionalRouter, serveOptions.ContextRoot)
	if log.Trace().Enabled() {
		log.Trace().
			Msg("Metrics handler registered")
	}

	// Context root router with OTEL
	contextRouter := parent.PathPrefix(serveOptions.ContextRoot).Subrouter()
	contextRouter.Use(middleware.TenantResolver)
	contextRouter.Use(middleware.RequestLogger)
	contextRouter.Path("/ui").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, serveOptions.ContextRoot+"/ui/", http.StatusTemporaryRedirect)
	})
	if log.Trace().Enabled() {
		log.Trace().
			Msg("Context router registered")
	}

	// Auth router
	if !oidcOptions.Disabled {
		authRouter := contextRouter.PathPrefix("/auth").Subrouter()
		authRouter.Use(middleware.InjectSession)
		if log.Trace().Enabled() {
			log.Trace().
				Msg("Auth router registered")
		}
		auth.IAMHandlers(authRouter, serveOptions.ContextRoot, relyingParty)
		if log.Trace().Enabled() {
			log.Trace().
				Msg("Auth handler registered")
		}
	}

	// Static content
	staticRouter := contextRouter.PathPrefix("/ui/").Subrouter()
	staticRouter.Use(middleware.InjectSession)
	staticRouter.Use(middleware.HttpSessionAuthenticationRequired)
	staticRouter.Use(middleware.HttpSessionInspectAndRenew)
	if log.Trace().Enabled() {
		log.Trace().
			Msg("Static router registered")
	}
	static.HandleStatic(staticRouter, serveOptions.ContextRoot, serveOptions.StaticPath)
	if log.Trace().Enabled() {
		log.Trace().
			Msg("Static handler registered")
	}

	// API router
	apiRouter := contextRouter.PathPrefix("/api").Subrouter()
	apiRouter.Use(mux.CORSMethodMiddleware(apiRouter))
	apiRouter.Use(middleware.JSONResponse)
	apiRouter.Use(middleware.PrometheusMiddleware)
	// TODO: gw oriented auth, inspect and renew
	// apiRouter.Use(middleware.MixedAuthenticationRequired)
	// apiRouter.Use(middleware.MixedInspectAndRenew)
	if log.Trace().Enabled() {
		log.Trace().
			Msg("API router registered")
	}

	return apiRouter
}

var routeCounter map[string]int = make(map[string]int)

func ProxyRoute(ctxRoot string, apiRouter *mux.Router, remoteRoute string) {

	after, found := strings.CutPrefix(remoteRoute, "unroute:")
	if found {
		resource := strings.TrimSuffix(after, ":")
		routeCounter[resource] -= 1
		if routeCounter[resource] <= 0 {
			removeProxy(ctxRoot, apiRouter, resource)
		}
		return
	}

	after, found = strings.CutPrefix(remoteRoute, "route:")
	if found {
		parts := strings.SplitAfterN(after, ":", 2)
		resource := strings.TrimSuffix(parts[0], ":")
		if routeCounter[resource] == 0 {
			createProxy(ctxRoot, apiRouter, after, parts, resource)
		}
		routeCounter[resource] += 1
		return
	}

	log.Warn().
		Str("route", remoteRoute).
		Msg("Invalid route protocol")
}

func removeProxy(ctxRoot string, apiRouter *mux.Router, resource string) {

	routeName := fmt.Sprintf("%s/api%s", ctxRoot, resource)
	route := apiRouter.Get(routeName)

	if route != nil {
		route.Handler(nil)

		log.Debug().
			Str("resource", resource).
			Any("endpoint", routeName).
			Msg("Dynamic endpoint unregistered")
	}
}

func createProxy(ctxRoot string, apiRouter *mux.Router, remoteRoute string, parts []string, resource string) {

	if len(parts) != 2 {
		log.Warn().
			Any("parts", parts).
			Msg("Invalid route resource")
		return
	}

	forward := parts[1]

	routeName := fmt.Sprintf("%s/api%s", ctxRoot, resource)
	route := apiRouter.Get(routeName)

	forwardURL, err := url.Parse(forward)
	if err != nil {
		log.Warn().
			Err(err).
			Str("route", remoteRoute).
			Msg("Invalid route URL")
		return
	}

	proxy := newReverseProxy(ctxRoot, resource, forwardURL)

	if route == nil || route.GetHandler() == nil {
		apiRouter.NewRoute().Name(routeName).Handler(proxy)

		log.Debug().
			Str("resource", resource).
			Str("forward", forward).
			Any("endpoint", routeName).
			Msg("Dynamic endpoint registered")
	}
}

func newReverseProxy(ctxRoot string, resource string, target *url.URL) *httputil.ReverseProxy {

	rewriteFun := func(r *httputil.ProxyRequest) {
		r.SetXForwarded()

		otelCarrier := propagation.HeaderCarrier(r.Out.Header)
		otel.GetTextMapPropagator().Inject(r.In.Context(), otelCarrier)
		log.Trace().
			Any("carrier", otelCarrier).
			Msg("OTEL propagation")

		r.Out = r.In
		r.Out = r.In.Clone(r.In.Context())

		tgtFunctionalRoot := strings.Replace(target.Path, resource, "", 1)
		r.Out.URL.Path = strings.Replace(r.In.URL.Path, ctxRoot+"/api", tgtFunctionalRoot, 1)

		r.Out.URL.Host = target.Host
		r.Out.URL.Scheme = target.Scheme
		r.Out.URL.User = target.User
		r.Out.Header = http.Header(otelCarrier)

		log.Trace().
			Any("in", r.In.Header).
			Any("out", r.Out.Header).
			Msg("...proxying...")
	}

	return &httputil.ReverseProxy{
		Rewrite: rewriteFun,
	}
}
