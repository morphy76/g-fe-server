package main

import (
	"context"
	"flag"
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
	"github.com/morphy76/g-fe-server/internal/http/session"
	"github.com/morphy76/g-fe-server/internal/logger"
	"github.com/morphy76/g-fe-server/internal/server"
)

const (
	envEnableTrace = "ENABLE_TRACE"
)

func main() {
	// Gather startup flags

	trace := flag.Bool("trace", false, "sets log level to trace. Environment: "+envEnableTrace)
	envTrace, found := os.LookupEnv(envEnableTrace)
	if found {
		*trace = strings.ToLower(envTrace) == "true"
	}

	serveOptionsBuilder := cli.ServeOptionsBuilder()
	sessionOptionsBuilder := cli.SessionOptionsBuilder()
	oidcOptionsBuilder := cli.OIDCOptionsBuilder()
	// OTelOptionsBuilder := cli.OTelOptionsBuilder()
	// dbOptionsBuilder := cli.DBOptionsBuilder()

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

	// OTelOptions, err := OTelOptionsBuilder()
	// if err != nil {
	// 	log.Error().
	// 		Err(err).
	// 		Msg("Error parsing OTel options")
	// 	flag.Usage()
	// 	os.Exit(1)
	// }

	oidcOptions, err := oidcOptionsBuilder()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error parsing oidc options")
		flag.Usage()
		os.Exit(1)
	}

	// dbOptions, err := dbOptionsBuilder()
	// if err != nil {
	// 	log.Error().
	// 		Err(err).
	// 		Msg("Error parsing db options")
	// 	flag.Usage()
	// 	os.Exit(1)
	// }

	startServer(
		serveOptions,
		sessionOptions,
		nil, // OTelOptions,
		oidcOptions,
		nil, // dbOptions,
		trace,
	)
}

func startServer(
	serveOptions *options.ServeOptions,
	sessionOptions *session.SessionOptions,
	otelOptions *options.OTelOptions,
	oidcOptions *auth.OIDCOptions,
	dbOptions *options.MongoDBOptions,
	trace *bool,
) {
	// manage termination criteria and channels
	srvErr := make(chan error, 1)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// HTTP session store
	sessionStore, err := session.CreateSessionStore(sessionOptions, serveOptions.ContextRoot)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error creating session store")
		return
	}

	// Server application context which provides the feServer instance and log facilities
	appContext, cancel := createAppContext(serveOptions, sessionOptions, sessionStore, oidcOptions, dbOptions, otelOptions, trace)
	bootLogger := logger.GetLogger(appContext, "feServer")

	// Server routes
	rootRouter := mux.NewRouter()
	server.Handler(appContext, rootRouter)
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
		srvErr <- server.ExtractFEServer(appContext).ListenAndServe(appContext, rootRouter)
	}()

	// Wait for termination signal
	for {
		select {
		// Termination OS signals
		case <-sigChan:
			cancel()
		// HTTP server error
		case err := <-srvErr:
			bootLogger.Err(err).Msg("Fail to start server")
			cancel()
		// Application context termination, triggered by OS signal or HTTP server error
		case <-appContext.Done():
			server.ExtractFEServer(appContext).Shutdown(appContext)
			return
		}
	}
}

func createAppContext(
	serveOpts *options.ServeOptions,
	sessionOptions *session.SessionOptions,
	sessionStore session.SessionStore,
	oidcOptions *auth.OIDCOptions,
	dbOptions *options.MongoDBOptions,
	otelOptions *options.OTelOptions,
	trace *bool,
) (context.Context, context.CancelFunc) {
	// as server application, the context is enriched with logger and server instance
	// the logger is structured, attributes are in the structure when resonable
	/*
		{
			"timing": {
				"timestamp": "",
				"since_start_ms": ""
			},
			"category": "",
			"owner": {
				"tenant_id": "",
				"subscription_id": "",
				"group_id": "",
				"system": false
			},
			"correlation": {
				"span_id": "",
				"trace_id": ""
			},
			"request": {
				"method": "",
				"path": "",
				"duration_ns": 0,
				"code": 200
			}
		}
	*/
	appContext := logger.InitLogger(context.Background(), trace)
	// as server application, the context is enriched with server instance
	// the server instance is the main entry point for the server application
	// it is used to start the server, register routes, and shutdown the server
	// it provides the HTTP server, with HTTP session, the OIDC client, and the database client
	appContext = server.NewFEServer(appContext, serveOpts, sessionOptions, sessionStore, oidcOptions, dbOptions, otelOptions)
	return context.WithCancel(appContext)
}
