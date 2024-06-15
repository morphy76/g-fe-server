package example

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/client/rs"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"

	"github.com/morphy76/g-fe-server/internal/db"
	example_http "github.com/morphy76/g-fe-server/internal/example/http"
	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/morphy76/g-fe-server/internal/http/handlers/health"
	"github.com/morphy76/g-fe-server/internal/http/handlers/metrics"
	"github.com/morphy76/g-fe-server/internal/http/middleware"
	"github.com/morphy76/g-fe-server/internal/serve"
)

func Handler(
	parent *mux.Router,
	app_context context.Context,
) {
	serveOptions := app_http.ExtractServeOptions(app_context)
	oidcOptions := app_http.ExtractOidcOptions(app_context)

	var relyingParty rp.RelyingParty
	var resourceServer rs.ResourceServer
	if !oidcOptions.Disabled {
		relyingParty = app_http.ExtractRelyingParty(app_context)
		resourceServer = app_http.ExtractOidcResource(app_context)
	}

	dbOptions := db.ExtractDbOptions(app_context)
	dbClient := db.ExtractDb(app_context)

	// Parent router
	parent.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			useRequest := r.WithContext(app_http.InjectServeOptions(r.Context(), serveOptions))
			useRequest = useRequest.WithContext(app_http.InjectOidcOptions(useRequest.Context(), oidcOptions))
			if !oidcOptions.Disabled {
				useRequest = useRequest.WithContext(app_http.InjectRelyingParty(useRequest.Context(), relyingParty))
				useRequest = useRequest.WithContext(app_http.InjectOidcResource(useRequest.Context(), resourceServer))
			}
			useRequest = useRequest.WithContext(db.InjectDbOptions(useRequest.Context(), dbOptions))
			useRequest = useRequest.WithContext(db.InjectDb(useRequest.Context(), dbClient))

			next.ServeHTTP(w, useRequest)
		})
	})

	parent.Use(otelmux.Middleware(serve.OTEL_EXAMPLE_NAME,
		otelmux.WithPublicEndpoint(),
		otelmux.WithPropagators(otel.GetTextMapPropagator()),
	))

	// Non functional router
	nonFunctionalRouter := parent.PathPrefix("/g").Subrouter()
	if log.Trace().Enabled() {
		log.Trace().Msg("Non functional router registered")
	}
	health.HealthHandlers(nonFunctionalRouter, app_context, db.CreateHealthCheck(dbOptions))
	if log.Trace().Enabled() {
		log.Trace().Msg("Health handler registered")
	}
	metrics.PrometheusHandlers(nonFunctionalRouter, serveOptions.ContextRoot)
	if log.Trace().Enabled() {
		log.Trace().Msg("Metrics handler registered")
	}

	// Context root router
	contextRouter := parent.PathPrefix(serveOptions.ContextRoot).Subrouter()
	contextRouter.Use(middleware.TenantResolver)
	contextRouter.Use(middleware.RequestLogger)
	if log.Trace().Enabled() {
		log.Trace().Msg("Context router registered")
	}

	// API router
	apiRouter := contextRouter.PathPrefix("/api").Subrouter()
	apiRouter.Use(mux.CORSMethodMiddleware(apiRouter))
	apiRouter.Use(middleware.JSONResponse)
	apiRouter.Use(middleware.PrometheusMiddleware)
	// TODO: bearer oriented auth, inspect and no renew
	// apiRouter.Use(middleware.BearerAuthenticationRequired)
	if log.Trace().Enabled() {
		log.Trace().Msg("API router registered")
	}

	// Domain functions
	example_http.ExampleHandlers(apiRouter, app_context)
}
