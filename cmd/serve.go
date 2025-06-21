package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/morphy76/g-fe-server/cmd/cli"
	"github.com/morphy76/g-fe-server/cmd/options"
	"github.com/morphy76/g-fe-server/internal/auth"
	"github.com/morphy76/g-fe-server/internal/http/handlers"
	"github.com/morphy76/g-fe-server/internal/logger"
	"github.com/morphy76/g-fe-server/internal/server"
)

const (
	envEnableTrace = "ENABLE_TRACE"
)

func main() {

	trace := flag.Bool("trace", false, "sets log level to trace. Environment: "+envEnableTrace)
	envTrace, found := os.LookupEnv(envEnableTrace)
	if found {
		*trace = strings.ToLower(envTrace) == "true"
	}

	serveOptionsBuilder := cli.ServeOptionsBuilder()
	sessionOptionsBuilder := cli.SessionOptionsBuilder()
	oidcOptionsBuilder := cli.OIDCOptionsBuilder()
	dbOptionsBuilder := cli.DBOptionsBuilder()
	OTelOptionsBuilder := cli.OTelOptionsBuilder()
	unleashOptionsBuilder := cli.UnleashOptionsBuilder()
	aiwOptionsBuilder := cli.AIWOptionsBuilder()

	help := flag.Bool("help", false, "prints help message")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	serveOptions, err := serveOptionsBuilder()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error parsing serve options")
		flag.Usage()
		os.Exit(1)
	}

	sessionOptions, err := sessionOptionsBuilder()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error parsing session options")
		flag.Usage()
		os.Exit(1)
	}
	sessionOptions.Path = serveOptions.ContextRoot

	oidcOptions, err := oidcOptionsBuilder()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error parsing oidc options")
		flag.Usage()
		os.Exit(1)
	}

	dbOptions, err := dbOptionsBuilder()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error parsing db options")
		flag.Usage()
		os.Exit(1)
	}

	oTelOptions, err := OTelOptionsBuilder()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error parsing OTel options")
		flag.Usage()
		os.Exit(1)
	}

	unleashOptions, err := unleashOptionsBuilder()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error parsing unleash options")
		flag.Usage()
		os.Exit(1)
	}

	AIWOptions, err := aiwOptionsBuilder()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error parsing AIW options")
		flag.Usage()
		os.Exit(1)
	}

	httpOptions := &options.HTTPOptions{
		ServeOptions:   serveOptions,
		SessionOptions: sessionOptions,
	}

	integrationOptions := &options.IntegrationOptions{
		DBOptions:      dbOptions,
		OTelOptions:    oTelOptions,
		UnleashOptions: unleashOptions,
		AIWOptions:     AIWOptions,
	}

	err = startServer(
		httpOptions,
		oidcOptions,
		integrationOptions,
		trace,
	)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error starting server")
		os.Exit(1)
	}
}

func startServer(
	httpOptions *options.HTTPOptions,
	oidcOptions *auth.OIDCOptions,
	integrationOptions *options.IntegrationOptions,
	trace *bool,
) error {
	// manage termination criteria and channels
	srvErr := make(chan error, 1)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Server application context which provides the feServer instance and log facilities
	appContext, cancel, err := createAppContext(
		httpOptions,
		oidcOptions,
		integrationOptions,
		trace,
	)
	if err != nil {
		return fmt.Errorf("failed to create application context: %w", err)
	}

	bootLogger := logger.GetLogger(appContext, "feServer")

	// Server routes
	rootRouter := mux.NewRouter()
	handlers.Handler(appContext, rootRouter)
	events := zerolog.Arr()
	rootRouter.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		if len(route.GetName()) > 0 {
			events.Str(route.GetName())
		}
		return nil
	})
	bootLogger.Info().Array("endpoints", events).Msg("Endpoint registered")

	// Start the HTTP server
	go func() {
		feServer, err := server.ExtractFEServer(appContext)
		if err != nil {
			srvErr <- fmt.Errorf("failed to extract FEServer from context: %w", err)
			return
		}
		srvErr <- feServer.ListenAndServe(appContext, rootRouter)
	}()

	// Wait for termination signal
	for {
		select {
		// Termination OS signals
		case <-sigChan:
			cancel()
		// HTTP server error
		case err := <-srvErr:
			bootLogger.Err(err).Msg("Server error")
			cancel()
		// Application context termination, triggered by OS signal or HTTP server error
		case <-appContext.Done():
			feServer, err := server.ExtractFEServer(appContext)
			if err != nil {
				bootLogger.Err(err).Msg("Server termination error")
				return nil
			}
			feServer.Shutdown(appContext)
			return nil
		}
	}
}

func createAppContext(
	httpOptions *options.HTTPOptions,
	oidcOptions *auth.OIDCOptions,
	integrationOptions *options.IntegrationOptions,
	trace *bool,
) (context.Context, context.CancelFunc, error) {
	appContext := logger.InitLogger(context.Background(), trace)
	appContext, err := server.NewFEServer(
		appContext,
		httpOptions,
		oidcOptions,
		integrationOptions,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create FEServer: %w", err)
	}

	rv, cancelFn := context.WithCancel(appContext)
	return rv, cancelFn, nil
}
