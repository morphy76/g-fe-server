package server

import (
	"context"
	"errors"
	"fmt"
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

// ErrFEServerNotFound is returned when the FEServer is not found in the context
var ErrFEServerNotFound = errors.New("FEServer not found in context")

// FEServer is a simple struct that represents an event bus
type FEServer struct {
	// UID is a unique identifier for the server instance
	UID string

	// ServeOpts contains options for serving the application
	ServeOpts *options.ServeOptions

	// SessionName is the name of the session used for storing user data
	SessionName string
	// SessionStore is the session store used for managing user sessions
	SessionStore sessions.Store
	// SessionOptions contains options for session management
	SessionOptions *session.SessionOptions

	// MongoClient is the MongoDB client used for database operations
	MongoClient *mongo.Client

	// RelyingParty is the OIDC relying party client used for authentication
	RelayingParty rp.RelyingParty
	// ResourceServer is the OIDC resource server client used for authorization
	ResourceServer rs.ResourceServer

	// ServiceName is the name of the service, used for logging and tracing
	ServiceName string

	// ShutdownFn is a list of functions to be called when shutting down the server
	ShutdownFn []func() error

	// HealthChecksFn is a list of additional health check functions
	HealthChecksFn []health.AdditionalCheckFn

	// AIWfacade is the AIW facade used for interacting with the AIW service
	AIWfacade *aiw.AIWFacade

	featureEnabled bool
}

// ExtractFEServer returns the FEServer from the context
func ExtractFEServer(ctx context.Context) (*FEServer, error) {
	rv := ctx.Value(appModelCtxKey)
	if rv == nil {
		return nil, ErrFEServerNotFound
	}
	return rv.(*FEServer), nil
}

// InjectFEServer adds the FEServer to the context
func InjectFEServer(ctx context.Context, appContext context.Context) (context.Context, error) {
	feServer, err := ExtractFEServer(appContext)
	if err != nil {
		return nil, fmt.Errorf("failed to extract FEServer from app context: %w", err)
	}
	return context.WithValue(ctx, appModelCtxKey, feServer), nil
}

// NewFEServer creates a Context with a new EventBus
func NewFEServer(
	appContext context.Context,
	serveOpts *options.ServeOptions,
	sessionOptions *session.SessionOptions,
	oidcOptions *auth.OIDCOptions,
	integrationsOptions *options.IntegrationOptions,
) (context.Context, error) {

	feServer := &FEServer{
		UID:         uuid.New().String(),
		ServeOpts:   serveOpts,
		ServiceName: integrationsOptions.OTelOptions.ServiceName,
		ShutdownFn:  make([]func() error, 0),

		featureEnabled: integrationsOptions.UnleashOptions.Enabled,
	}

	err := bindInfrastructuralDependencies(
		feServer,
		serveOpts,
		oidcOptions,
		sessionOptions,
		integrationsOptions,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to bind infrastructural dependencies: %w", err)
	}

	err = addHealthChecks(feServer, integrationsOptions.DBOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to add health checks: %w", err)
	}

	return context.WithValue(appContext, appModelCtxKey, feServer), nil
}

// ListenAndServe starts the server
func (feServer *FEServer) ListenAndServe(ctx context.Context, rootRouter *mux.Router) error {
	feLogger := logger.GetLogger(ctx, "feServer")

	feLogger.Info().
		Dict("serve_opts", zerolog.Dict().
			Str("protocol", feServer.ServeOpts.Protocol).
			Str("host", feServer.ServeOpts.Host).
			Str("port", feServer.ServeOpts.Port).
			Str("ctx", feServer.ServeOpts.ContextRoot).
			Str("serving", feServer.ServeOpts.StaticPath)).
		Dict("aiw", zerolog.Dict().
			Str("fqdn", feServer.AIWfacade.AIWOptions.FQDN)).
		Msg("Server started")

	if feServer.ServeOpts.Protocol == "https" {
		return http.ListenAndServeTLS(
			feServer.ServeOpts.Host+":"+feServer.ServeOpts.Port,
			feServer.ServeOpts.CertFile,
			feServer.ServeOpts.KeyFile,
			rootRouter,
		)
	} else {
		return http.ListenAndServe(
			feServer.ServeOpts.Host+":"+feServer.ServeOpts.Port,
			rootRouter,
		)
	}

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

// IsFeatureEnabled checks if a feature is enabled using Unleash
func (feServer *FEServer) IsFeatureEnabled(feature string, opts ...unleash.FeatureOption) bool {
	if !feServer.featureEnabled {
		return true
	}
	return unleash.IsEnabled(feature, opts...)
}
