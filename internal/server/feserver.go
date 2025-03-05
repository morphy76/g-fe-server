package server

import (
	"context"
	"net/http"

	"github.com/Unleash/unleash-client-go/v4"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/morphy76/g-fe-server/cmd/options"
	"github.com/morphy76/g-fe-server/internal/aiw"
	"github.com/morphy76/g-fe-server/internal/auth"
	"github.com/morphy76/g-fe-server/internal/common"
	"github.com/morphy76/g-fe-server/internal/common/health"
	"github.com/morphy76/g-fe-server/internal/http/session"
	"github.com/morphy76/g-fe-server/internal/logger"
	"github.com/rs/zerolog"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/client/rs"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/google/uuid"
)

const appModelCtxKey common.CtxKey = "App"

// FEServer is a simple struct that represents an event bus
type FEServer struct {
	UID string

	ServeOpts *options.ServeOptions

	SessionName  string
	SessionStore sessions.Store

	MongoClient *mongo.Client

	RelayingParty  rp.RelyingParty
	ResourceServer rs.ResourceServer

	ServiceName string
	ShutdownFn  []func() error

	HealthChecksFn []health.AdditionalCheckFn

	featureEnabled bool

	AIWfacade *aiw.AIWFacade
}

// ExtractFEServer returns the FEServer from the context
func ExtractFEServer(ctx context.Context) *FEServer {
	return ctx.Value(appModelCtxKey).(*FEServer)
}

// InjectFEServer adds the FEServer to the context
func InjectFEServer(ctx context.Context, appContext context.Context) context.Context {
	feServer := ExtractFEServer(appContext)
	return context.WithValue(ctx, appModelCtxKey, feServer)
}

// NewFEServer creates a Context with a new EventBus
func NewFEServer(
	ctx context.Context,
	serveOpts *options.ServeOptions,
	sessionOptions *session.SessionOptions,
	oidcOptions *auth.OIDCOptions,
	dbOptions *options.MongoDBOptions,
	otelOptions *options.OTelOptions,
	unleashOptions *options.UnleashOptions,
	aiwOptions *options.AIWOptions,
) context.Context {

	feServer := &FEServer{
		UID:         uuid.New().String(),
		ServeOpts:   serveOpts,
		ServiceName: otelOptions.ServiceName,
		ShutdownFn:  make([]func() error, 0),

		featureEnabled: unleashOptions.Enabled,
	}

	err := bindInfrastructuralDependencies(
		feServer,
		serveOpts,
		oidcOptions,
		sessionOptions,
		dbOptions,
		otelOptions,
		unleashOptions,
		aiwOptions,
	)
	if err != nil {
		panic(err)
	}

	err = addHealthChecks(feServer, dbOptions)
	if err != nil {
		panic(err)
	}

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
		Dict("aiw", zerolog.Dict().
			Str("fqdn", feServer.AIWfacade.AIWOptions.FQDN)).
		Msg("Server started")

	return http.ListenAndServe(feServer.ServeOpts.Host+":"+feServer.ServeOpts.Port, rootRouter)
}

// Shutdown stops the server
func (feServer *FEServer) Shutdown(ctx context.Context) {
	feLogger := logger.GetLogger(ctx, "feServer")
	for _, fn := range feServer.ShutdownFn {
		if err := fn(); err != nil {
			feLogger.Error().Err(err).Msg("Error shutting down")
		}
	}
	feLogger.Info().Msg("Server stopped")
}

func (feServer *FEServer) IsFeatureEnabled(feature string, opts ...unleash.FeatureOption) bool {
	if !feServer.featureEnabled {
		return true
	}
	return unleash.IsEnabled(feature, opts...)
}
