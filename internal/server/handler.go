package server

import (
	"context"
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

	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/morphy76/g-fe-server/internal/http/handlers/auth"
	"github.com/morphy76/g-fe-server/internal/http/handlers/health"
	"github.com/morphy76/g-fe-server/internal/http/handlers/metrics"
	"github.com/morphy76/g-fe-server/internal/http/handlers/static"
	"github.com/morphy76/g-fe-server/internal/http/middleware"
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
		log.Trace().Msg("Non functional router registered")
	}
	health.HealthHandlers(nonFunctionalRouter, app_context)
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
		http.Redirect(w, r, serveOptions.ContextRoot+"/ui/", http.StatusTemporaryRedirect)
	})
	if log.Trace().Enabled() {
		log.Trace().Msg("Context router registered")
	}

	// Auth router
	if !oidcOptions.Disabled {
		authRouter := contextRouter.PathPrefix("/auth").Subrouter()
		authRouter.Use(middleware.InjectSession)
		if log.Trace().Enabled() {
			log.Trace().Msg("Auth router registered")
		}
		auth.IAMHandlers(authRouter, serveOptions.ContextRoot, relyingParty)
		if log.Trace().Enabled() {
			log.Trace().Msg("Auth handler registered")
		}
	}

	// Static content
	staticRouter := contextRouter.PathPrefix("/ui/").Subrouter()
	staticRouter.Use(middleware.InjectSession)
	staticRouter.Use(middleware.AuthenticationRequired)
	staticRouter.Use(middleware.InspectAndRenew)
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

	contextRouter.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		if len(route.GetName()) > 0 {
			router.Use(otelmux.Middleware(route.GetName(),
				otelmux.WithPublicEndpoint(),
				otelmux.WithPropagators(otel.GetTextMapPropagator()),
			))
		}
		return nil
	})

	return apiRouter
}

func ProxyRoute(apiRouter *mux.Router, remoteRoute string) {
	after, found := strings.CutPrefix(remoteRoute, "route:")
	if found {
		createProxy(apiRouter, after)
	} else {
		after, found = strings.CutPrefix(remoteRoute, "unroute:")
		if found {
			removeProxy(apiRouter, after)
		} else {
			log.Warn().Str("route", remoteRoute).Msg("Invalid route protocol")
		}
	}
}

func removeProxy(apiRouter *mux.Router, remoteRoute string) {

	parts := strings.SplitAfterN(remoteRoute, ":", 2)
	if len(parts) < 1 {
		log.Warn().Any("parts", parts).Msg("Invalid route resource")
		return
	}

	route := apiRouter.Get(parts[0])
	if route != nil {
		route.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "", http.StatusNotFound)
		})
	}
}

func createProxy(apiRouter *mux.Router, remoteRoute string) {

	parts := strings.SplitAfterN(remoteRoute, ":", 2)
	if len(parts) != 2 {
		log.Warn().Any("parts", parts).Msg("Invalid route resource")
		return
	}

	resource := parts[0]
	forward := parts[1]

	route := apiRouter.Get(resource)

	log.Trace().
		Str("resource", resource).
		Str("forward", forward).
		Msg("Proxying route")

	forwardURL, err := url.Parse(forward)
	if err != nil {
		log.Warn().Err(err).Str("route", remoteRoute).Msg("Invalid route URL")
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(forwardURL)

	if route == nil {
		apiRouter.NewRoute().Name(resource).PathPrefix(resource).Handler(proxy)
	} else {
		route.Handler(proxy)
	}
}
