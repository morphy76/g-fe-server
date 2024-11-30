package server

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/morphy76/g-fe-server/internal/common"
	"github.com/morphy76/g-fe-server/internal/logger"
	"github.com/morphy76/g-fe-server/internal/options"
	"github.com/morphy76/g-fe-server/internal/serve"
	"github.com/rs/zerolog"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/client/rs"

	"github.com/google/uuid"
)

const appModelCtxKey common.CtxKey = "App"

// FEServer is a simple struct that represents an event bus
type FEServer struct {
	UID string

	ServeOpts    *options.ServeOptions
	SessionStore sessions.Store
	DBOpts       *options.MongoDBOptions
	OTelOpts     *options.OTelOptions

	RelayingParty  rp.RelyingParty
	ResourceServer rs.ResourceServer

	OtelShutdownFn func() error
}

// ExtractFEServer returns the FEServer from the context
func ExtractFEServer(ctx context.Context) *FEServer {
	return ctx.Value(appModelCtxKey).(*FEServer)
}

// NewFEServer creates a Context with a new EventBus
func NewFEServer(
	ctx context.Context,
	serveOpts *options.ServeOptions,
	sessionStore sessions.Store,
	oidcOptions *options.OIDCOptions,
	dbOptions *options.MongoDBOptions,
	otelOptions *options.OTelOptions,
) context.Context {

	otelShutdown, err := serve.SetupOTelSDK(otelOptions)
	if err != nil {
		panic(err)
	}

	feServer := &FEServer{
		UID: uuid.New().String(),

		ServeOpts:    serveOpts,
		SessionStore: sessionStore,
		DBOpts:       dbOptions,
		OTelOpts:     otelOptions,

		OtelShutdownFn: otelShutdown,
	}

	rp, err := serve.SetupOIDC(serveOpts, oidcOptions)
	if err != nil {
		panic(err)
	}
	feServer.RelayingParty = rp

	rs, err := rs.NewResourceServerClientCredentials(context.Background(), oidcOptions.Issuer, oidcOptions.ClientID, oidcOptions.ClientSecret)
	if err != nil {
		panic(err)
	}
	feServer.ResourceServer = rs

	return context.WithValue(ctx, appModelCtxKey, feServer)
}

// ListenAndServe starts the server
func (feServer *FEServer) ListenAndServe(ctx context.Context, rootRouter *mux.Router) error {
	feLogger := logger.GetLogger(ctx, "feServer")

	feLogger.Info().
		Dict("serve_opts", zerolog.Dict().
			Str("host", feServer.ServeOpts.Host).
			Str("port", feServer.ServeOpts.Port).
			Str("ctx", feServer.ServeOpts.ContextRoot).
			Str("serving", feServer.ServeOpts.StaticPath)).
		Dict(("oidc_opts"), zerolog.Dict().
			Str("issuer", feServer.RelayingParty.Issuer())).
		Dict("db_opts", zerolog.Dict().
			Str("url", feServer.DBOpts.URL)).
		Dict("otel_opts", zerolog.Dict().
			Bool("enabled", feServer.OTelOpts.Enabled).
			Str("url", feServer.OTelOpts.URL)).
		Msg("Server started")

	return http.ListenAndServe(feServer.ServeOpts.Host+":"+feServer.ServeOpts.Port, rootRouter)
}

// Shutdown stops the server
func (feServer *FEServer) Shutdown(ctx context.Context) {
	feLogger := logger.GetLogger(ctx, "feServer")
	if feServer.OtelShutdownFn != nil {
		if err := feServer.OtelShutdownFn(); err != nil {
			feLogger.Error().Err(err).Msg("Error shutting down opentelemetry")
		}
	}
	feLogger.Info().Msg("Server stopped")
}
