package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Unleash/unleash-client-go/v4"
	"github.com/morphy76/g-fe-server/cmd/options"
	"github.com/morphy76/g-fe-server/internal/aiw"
	"github.com/morphy76/g-fe-server/internal/auth"
	"github.com/morphy76/g-fe-server/internal/common/health"
	"github.com/morphy76/g-fe-server/internal/db"
	"github.com/morphy76/g-fe-server/internal/http/session"
	"github.com/morphy76/g-fe-server/internal/otel"
	"github.com/zitadel/oidc/v3/pkg/client/rs"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func bindInfrastructuralDependencies(
	feServer *FEServer,
	serveOpts *options.ServeOptions,
	oidcOptions *auth.OIDCOptions,
	sessionOptions *session.SessionOptions,
	integrationsOptions *options.IntegrationOptions,
) error {

	otelShutdown, err := otel.SetupOTelSDK(integrationsOptions.OTelOptions)
	if err != nil {
		return fmt.Errorf("failed to setup OTel: %w", err)
	}
	if otelShutdown != nil {
		feServer.ShutdownFn = append(feServer.ShutdownFn, otelShutdown)
	}

	err = bindOIDC(feServer, serveOpts, oidcOptions)
	if err != nil {
		return fmt.Errorf("failed to bind OIDC: %w", err)
	}

	err = bindSessionStore(feServer, serveOpts, sessionOptions, integrationsOptions.DBOptions)
	if err != nil {
		return fmt.Errorf("failed to bind session store: %w", err)
	}

	err = bindMongoDB(feServer, err, integrationsOptions.DBOptions, integrationsOptions.OTelOptions.Enabled)
	if err != nil {
		return fmt.Errorf("failed to bind MongoDB: %w", err)
	}

	err = bindUnleash(integrationsOptions.UnleashOptions)
	if err != nil {
		return fmt.Errorf("failed to bind Unleash: %w", err)
	}

	err = bindAIW(feServer, integrationsOptions.AIWOptions)
	if err != nil {
		return fmt.Errorf("failed to bind AIW: %w", err)
	}

	return nil
}

func bindAIW(feServer *FEServer, aiwOptions *options.AIWOptions) error {
	aiwFacade := &aiw.AIWFacade{
		AIWOptions: aiwOptions,
		HttpClient: instrumentNewHTTPClient(),
	}
	feServer.AIWfacade = aiwFacade

	return nil
}

func bindUnleash(unleashOptions *options.UnleashOptions) error {
	if !unleashOptions.Enabled {
		return nil
	}

	err := unleash.Initialize(
		unleash.WithHttpClient(instrumentUnleashHTTPClient()),
		// unleash.WithListener(unleash.DebugListener{}),
		unleash.WithAppName(unleashOptions.AppName),
		unleash.WithUrl(unleashOptions.URL),
		unleash.WithCustomHeaders(http.Header{"Authorization": {unleashOptions.Token}}),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize Unleash: %w", err)
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
	rp, err := auth.SetupOIDC(
		serveOpts.Protocol,
		serveOpts.Host,
		serveOpts.Port,
		serveOpts.ContextRoot,
		oidcOptions,
	)
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

func instrumentUnleashHTTPClient() *http.Client {
	transport := otelhttp.NewTransport(http.DefaultTransport)
	client := &http.Client{
		Transport: transport,
	}
	return client
}
