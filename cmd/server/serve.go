package main

import (
	"context"
	"flag"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/quasoft/memstore"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/morphy76/g-fe-server/cmd/cli"
	"github.com/morphy76/g-fe-server/internal/logger"
	"github.com/morphy76/g-fe-server/internal/options"
	"github.com/morphy76/g-fe-server/internal/server"
)

func main() {

	trace := flag.Bool("trace", false, "sets log level to trace")

	serveOptionsBuilder := cli.ServeOptionsBuilder()
	// otelOptionsBuilder := cli.OtelOptionsBuilder()
	oidcOptionsBuilder := cli.OIDCOptionsBuilder()

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

	// otelOptions, err := otelOptionsBuilder()
	// if err != nil {
	// 	log.Error().
	// 		Err(err).
	// 		Msg("Error parsing otel options")
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

	startServer(
		serveOptions,
		// otelOptions,
		oidcOptions,
		trace,
	)
}

func startServer(
	serveOptions *options.ServeOptions,
	// otelOptions *options.OtelOptions,
	oidcOptions *options.OIDCOptions,
	trace *bool,
) {
	sessionStore := createSessionStore(serveOptions)

	appContext, cancel := createAppContext(serveOptions, sessionStore, oidcOptions, trace)
	// appContext, cancel := createAppContext(serveOptions, sessionStore, oidcOptions, otelOptions, trace)
	bootLogger := logger.GetLogger(appContext, "feServer")
	defer cancel()

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

	srvErr := make(chan error, 1)
	go func() {
		srvErr <- server.ExtractFEServer(appContext).ListenAndServe(appContext, rootRouter)
	}()

	select {
	case err := <-srvErr:
		bootLogger.Info().
			Err(err).
			Msg("Server stopped")
	case <-appContext.Done():
		bootLogger.Info().
			Msg("Server stopped")
		cancel()
	}
}

func createSessionStore(serveOptions *options.ServeOptions) sessions.Store {
	// TODO from memstore to https://github.com/kidstuff/mongostore
	sessionStore := memstore.NewMemStore([]byte(serveOptions.SessionKey))
	sessionStore.Options = &sessions.Options{
		Path:     serveOptions.ContextRoot,
		MaxAge:   serveOptions.SessionMaxAge,
		HttpOnly: serveOptions.SessionHttpOnly,
		Domain:   serveOptions.SessionDomain,
		Secure:   serveOptions.SessionSecureCookies,
		SameSite: serveOptions.SessionSameSite,
	}
	return sessionStore
}

func createAppContext(
	serveOpts *options.ServeOptions,
	sessionStore sessions.Store,
	oidcOptions *options.OIDCOptions,
	// otelOptions *options.OtelOptions,
	trace *bool,
) (context.Context, context.CancelFunc) {
	appContext := logger.InitLogger(context.Background(), trace)
	appContext = server.NewFEServer(appContext, serveOpts, sessionStore, oidcOptions)
	// appContext = server.NewFEServer(appContext, serveOpts, sessionStore, oidcOptions, otelOptions)
	return context.WithCancel(appContext)
}
