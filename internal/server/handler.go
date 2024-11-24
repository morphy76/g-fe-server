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
	parent *mux.Router,
) *mux.Router {
	// sessionStore := app_http.ExtractSessionStore(app_context)
	// oidcOptions := app_http.ExtractOidcOptions(app_context)

	// var relyingParty rp.RelyingParty
	// var resourceServer rs.ResourceServer
	// if !oidcOptions.Disabled {
	// 	relyingParty = app_http.ExtractRelyingParty(app_context)
	// 	resourceServer = app_http.ExtractOidcResource(app_context)
	// }

	// Parent router
	// parent.Use(otelmux.Middleware(serve.OTEL_GW_NAME,
	// 	otelmux.WithPublicEndpoint(),
	// 	otelmux.WithPropagators(otel.GetTextMapPropagator()),
	// 	otelmux.WithTracerProvider(otel.GetTracerProvider()),
	// ))

	feServer := ExtractFEServer(appContext)
	routerLog := logger.GetLogger(appContext, "router")

	parent.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// useRequest := r.WithContext(app_http.InjectSessionStore(r.Context(), sessionStore))
			// useRequest = useRequest.WithContext(app_http.InjectOidcOptions(useRequest.Context(), oidcOptions))
			// if !oidcOptions.Disabled {
			// 	useRequest = useRequest.WithContext(app_http.InjectRelyingParty(useRequest.Context(), relyingParty))
			// 	useRequest = useRequest.WithContext(app_http.InjectOidcResource(useRequest.Context(), resourceServer))
			// }

			// next.ServeHTTP(w, useRequest)
			next.ServeHTTP(w, r)
		})
	})

	// Non functional router
	nonFunctionalRouter := parent.PathPrefix(feServer.NonFunctionalRoot).Subrouter()
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Non functional router registered")
	}
	// Add additional checks to test mongodb
	health.Handlers(appContext, nonFunctionalRouter, feServer.ServeOpts)
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("Health handler registered")
	}
	// metrics.PrometheusHandlers(nonFunctionalRouter, serveOptions.ContextRoot)
	// if log.Trace().Enabled() {
	// 	log.Trace().
	// 		Msg("Metrics handler registered")
	// }

	// Context root router with OTEL
	contextRouter := parent.PathPrefix(feServer.ServeOpts.ContextRoot).Subrouter()
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
	// if !oidcOptions.Disabled {
	// 	authRouter := contextRouter.PathPrefix("/auth").Subrouter()
	// 	authRouter.Use(middleware.InjectSession)
	// 	if log.Trace().Enabled() {
	// 		log.Trace().
	// 			Msg("Auth router registered")
	// 	}
	// 	auth.IAMHandlers(authRouter, serveOptions.ContextRoot, relyingParty)
	// 	if log.Trace().Enabled() {
	// 		log.Trace().
	// 			Msg("Auth handler registered")
	// 	}
	// }

	// Static content
	staticRouter := contextRouter.PathPrefix("/ui/").Subrouter()
	// staticRouter.Use(middleware.InjectSession)
	// staticRouter.Use(middleware.HttpSessionAuthenticationRequired)
	// staticRouter.Use(middleware.HttpSessionInspectAndRenew)
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
	apiRouter.Use(mux.CORSMethodMiddleware(apiRouter))
	apiRouter.Use(middleware.JSONResponse)
	// apiRouter.Use(middleware.PrometheusMiddleware)
	// TODO: gw oriented auth, inspect and renew
	// apiRouter.Use(middleware.MixedAuthenticationRequired)
	// apiRouter.Use(middleware.MixedInspectAndRenew)
	if routerLog.Trace().Enabled() {
		routerLog.Trace().
			Msg("API router registered")
	}

	return apiRouter
}
