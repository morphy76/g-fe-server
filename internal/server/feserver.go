package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/morphy76/g-fe-server/cmd/options"
	"github.com/morphy76/g-fe-server/internal/aiw"
	"github.com/morphy76/g-fe-server/internal/auth"
	"github.com/morphy76/g-fe-server/internal/common"
	"github.com/morphy76/g-fe-server/internal/common/health"
	"github.com/morphy76/g-fe-server/internal/db"
	"github.com/morphy76/g-fe-server/internal/http/session"
	"github.com/morphy76/g-fe-server/internal/logger"
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

	ServiceName string
	ShutdownFn  []func() error

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

	feServer := &FEServer{
		UID:               uuid.New().String(),
		ServeOpts:         serveOpts,
		BackendHTTPClient: instrumentNewHTTPClient(),
		ServiceName:       otelOptions.ServiceName,
		ShutdownFn:        make([]func() error, 0),
	}

	otelShutdown, err := SetupOTelSDK(otelOptions)
	if err != nil {
		panic(err)
	}
	if otelShutdown != nil {
		feServer.ShutdownFn = append(feServer.ShutdownFn, otelShutdown)
	}

	err = bindInfrastructuralDependencies(
		feServer,
		serveOpts,
		oidcOptions,
		sessionOptions,
		dbOptions,
		otelOptions,
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
	for _, fn := range feServer.ShutdownFn {
		if err := fn(); err != nil {
			feLogger.Error().Err(err).Msg("Error shutting down")
		}
	}
	feLogger.Info().Msg("Server stopped")
}

func (feServer *FEServer) GetAIWFacade() *aiw.AIWFacade {
	return &aiw.AIWFacade{
		HttpClient: feServer.BackendHTTPClient,
	}
}

func bindInfrastructuralDependencies(
	feServer *FEServer,
	serveOpts *options.ServeOptions,
	oidcOptions *auth.OIDCOptions,
	sessionOptions *session.SessionOptions,
	dbOptions *options.MongoDBOptions,
	otelOptions *options.OTelOptions,
) error {
	err := bindOIDC(feServer, serveOpts, oidcOptions)
	if err != nil {
		return fmt.Errorf("failed to bind OIDC: %w", err)
	}

	err = bindSessionStore(feServer, serveOpts, sessionOptions, dbOptions)
	if err != nil {
		return fmt.Errorf("failed to bind session store: %w", err)
	}

	err = bindMongoDB(feServer, err, dbOptions, otelOptions.Enabled)
	if err != nil {
		return fmt.Errorf("failed to bind MongoDB: %w", err)
	}

	return nil
}

func addHealthChecks(feServer *FEServer, dbOptions *options.MongoDBOptions) error {
	feServer.HealthChecksFn = make([]health.AdditionalCheckFn, 0)

	healthClient, err := db.NewClient(dbOptions, false)
	if err != nil {
		return err
	}

	feServer.HealthChecksFn = append(feServer.HealthChecksFn, db.CreateHealthCheck(healthClient))
	feServer.HealthChecksFn = append(feServer.HealthChecksFn, auth.CreateHealthCheck(feServer.RelayingParty))

	return nil
}

func bindMongoDB(feServer *FEServer, err error, dbOptions *options.MongoDBOptions, withMonitor bool) error {
	client, err := db.NewClient(dbOptions, withMonitor)
	if err != nil {
		return err
	}
	feServer.MongoClient = client
	shutdownFn := func() error {
		return client.Disconnect(context.Background())
	}
	feServer.ShutdownFn = append(feServer.ShutdownFn, shutdownFn)

	return nil
}

func bindSessionStore(
	feServer *FEServer,
	serveOpts *options.ServeOptions,
	sessionOptions *session.SessionOptions,
	dbOptions *options.MongoDBOptions,
) error {
	sessionStore, shutdownFn, err := session.CreateSessionStore(sessionOptions, dbOptions, serveOpts.ContextRoot)
	if err != nil {
		return err
	}
	feServer.SessionName = sessionOptions.SessionName
	feServer.SessionStore = sessionStore
	if shutdownFn != nil {
		feServer.ShutdownFn = append(feServer.ShutdownFn, shutdownFn)
	}

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
