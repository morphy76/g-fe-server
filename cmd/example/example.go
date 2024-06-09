package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/morphy76/g-fe-server/cmd/cli"
	"github.com/morphy76/g-fe-server/internal/db"
	"github.com/morphy76/g-fe-server/internal/example"
	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/morphy76/g-fe-server/internal/options"
)

func main() {

	zerolog.TimeFieldFormat = time.RFC3339
	trace := flag.Bool("trace", false, "sets log level to trace")

	serveOptionsBuilder := cli.ServeOptionsBuilder()
	otelOptionsBuilder := cli.OtelOptionsBuilder()
	oidcOptionsBuilder := cli.OidcOptionsBuilder()
	dbOptionsBuilder := cli.DbOptionsBuilder()

	help := flag.Bool("help", false, "prints help message")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	if *trace {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}

	serveOptions, err := serveOptionsBuilder()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error parsing serve options")
		flag.Usage()
		os.Exit(1)
	}

	otelOptions, err := otelOptionsBuilder()
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error parsing otel options")
		flag.Usage()
		os.Exit(1)
	}

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

	startServer(
		serveOptions,
		otelOptions,
		oidcOptions,
		dbOptions,
	)
}

func startServer(
	serveOptions *options.ServeOptions,
	otelOptions *options.OtelOptions,
	oidcOptions *options.OidcOptions,
	dbOptions *options.DbOptions,
) {
	start := time.Now()

	initialContext, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	shutdown, err := cli.SetupOTEL(initialContext, otelOptions)
	defer shutdown()
	if err != nil {
		panic(err)
	}

	dbClient, err := db.NewClient(dbOptions)
	if err != nil {
		panic(err)
	}

	serverContext := app_http.InjectServeOptions(initialContext, serveOptions)
	oidOptionsContext := app_http.InjectOidcOptions(serverContext, oidcOptions)
	oidcContext := cli.CreateTheOIDCContext(oidOptionsContext, oidcOptions, serveOptions)
	finalContext := db.InjectDb(db.InjectDbOptions(oidcContext, dbOptions), dbClient)
	log.Trace().
		Msg("Application contextes ready")

	rootRouter := mux.NewRouter()
	example.Handler(rootRouter, finalContext)
	if log.Trace().Enabled() {
		rootRouter.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			if len(route.GetName()) > 0 {
				log.Trace().Str("endpoint", route.GetName()).Msg("Endpoint registered")
			}
			return nil
		})
	}

	srvErr := make(chan error, 1)
	go func() {
		srvErr <- http.ListenAndServe(fmt.Sprintf("%s:%s", serveOptions.Host, serveOptions.Port), rootRouter)
	}()

	log.Info().
		Str("host", serveOptions.Host).
		Str("port", serveOptions.Port).
		Str("ctx", serveOptions.ContextRoot).
		Str("serving", serveOptions.StaticPath).
		Int64("setup_ns", time.Since(start).Nanoseconds()).
		Msg("Server started")

	select {
	case err = <-srvErr:
		log.Info().
			Err(err).
			Msg("Server stopped")
	case <-finalContext.Done():
		log.Info().
			Msg("Server stopped")
		stop()
	}
}
