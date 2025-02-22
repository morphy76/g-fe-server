package server

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/morphy76/g-fe-server/cmd/options"
	"github.com/morphy76/g-fe-server/internal/auth"
	"github.com/morphy76/g-fe-server/internal/common"
	"github.com/morphy76/g-fe-server/internal/common/health"
	"github.com/morphy76/g-fe-server/internal/db"
	"github.com/morphy76/g-fe-server/internal/http/session"
	"github.com/morphy76/g-fe-server/internal/logger"
	"github.com/morphy76/g-fe-server/internal/serve"
	"github.com/rs/zerolog"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/client/rs"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/google/uuid"
)

const appModelCtxKey common.CtxKey = "App"

// FEServer is a simple struct that represents an event bus
type FEServer struct {
	UID string

	ServeOpts *options.ServeOptions

	SessionName  string
	SessionStore sessions.Store

	BackendHTTPClient *http.Client

	MongoClient *mongo.Client

	RelayingParty  rp.RelyingParty
	ResourceServer rs.ResourceServer

	ServiceName    string
	OtelShutdownFn func() error

	HealthChecksFn []health.AdditionalCheckFn
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
) context.Context {

	otelShutdown, err := serve.SetupOTelSDK(otelOptions)
	if err != nil {
		panic(err)
	}

	feServer := &FEServer{
		UID:               uuid.New().String(),
		ServeOpts:         serveOpts,
		BackendHTTPClient: instrumentNewHTTPClient(),
		ServiceName:       otelOptions.ServiceName,
		OtelShutdownFn:    otelShutdown,
	}

	err = bindInfrastructuralDependencies(
		feServer,
		serveOpts,
		oidcOptions,
		sessionOptions,
		dbOptions,
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
		Msg("Server started")

	return http.ListenAndServe(feServer.ServeOpts.Host+":"+feServer.ServeOpts.Port, rootRouter)
}

// Shutdown stops the server
func (feServer *FEServer) Shutdown(ctx context.Context) {
	feLogger := logger.GetLogger(ctx, "feServer")
	err := feServer.MongoClient.Disconnect(context.Background())
	if err != nil {
		feLogger.Error().Err(err).Msg("Error disconnecting from MongoDB")
	}
	if feServer.OtelShutdownFn != nil {
		if err := feServer.OtelShutdownFn(); err != nil {
			feLogger.Error().Err(err).Msg("Error shutting down opentelemetry")
		}
	}
	feLogger.Info().Msg("Server stopped")
}

func bindInfrastructuralDependencies(
	feServer *FEServer,
	serveOpts *options.ServeOptions,
	oidcOptions *auth.OIDCOptions,
	sessionOptions *session.SessionOptions,
	dbOptions *options.MongoDBOptions,
) error {
	err := bindOIDC(feServer, serveOpts, oidcOptions)
	if err != nil {
		return err
	}

	err = bindSessionStore(feServer, serveOpts, sessionOptions)
	if err != nil {
		return err
	}

	feServer.MongoClient, err = db.NewClient(dbOptions)
	if err != nil {
		panic(err)
	}

	return nil
}

func addHealthChecks(feServer *FEServer, dbOptions *options.MongoDBOptions) error {
	feServer.HealthChecksFn = make([]health.AdditionalCheckFn, 0)

	healthClient, err := db.NewClient(dbOptions)
	if err != nil {
		return err
	}

	feServer.HealthChecksFn = append(feServer.HealthChecksFn, db.CreateHealthCheck(healthClient))
	feServer.HealthChecksFn = append(feServer.HealthChecksFn, auth.CreateHealthCheck(feServer.RelayingParty))

	return nil
}

func bindSessionStore(
	feServer *FEServer,
	serveOpts *options.ServeOptions,
	sessionOptions *session.SessionOptions,
) error {
	sessionStore, err := session.CreateSessionStore(sessionOptions, serveOpts.ContextRoot)
	if err != nil {
		return err
	}
	feServer.SessionName = sessionOptions.SessionName
	feServer.SessionStore = sessionStore

	return nil
}

func bindOIDC(
	feServer *FEServer,
	serveOpts *options.ServeOptions,
	oidcOptions *auth.OIDCOptions,
) error {
	rp, err := auth.SetupOIDC(serveOpts, oidcOptions)
	if err != nil {
		return err
	}
	feServer.RelayingParty = rp

	rs, err := rs.NewResourceServerClientCredentials(context.Background(), oidcOptions.Issuer, oidcOptions.ClientID, oidcOptions.ClientSecret)
	if err != nil {
		return err
	}
	feServer.ResourceServer = rs

	return nil
}

func instrumentNewHTTPClient() *http.Client {
	transport := otelhttp.NewTransport(http.DefaultTransport)
	client := &http.Client{
		Transport: transport,
	}
	return client
}
