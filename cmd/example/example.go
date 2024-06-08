package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/quasoft/memstore"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/zitadel/oidc/v3/pkg/client/rs"

	"github.com/morphy76/g-fe-server/cmd/cli"
	"github.com/morphy76/g-fe-server/internal/db"
	"github.com/morphy76/g-fe-server/internal/example"
	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/morphy76/g-fe-server/internal/options"
	"github.com/morphy76/g-fe-server/internal/serve"
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

	sessionStore := memstore.NewMemStore([]byte(serveOptions.SessionKey))
	sessionStore.Options = &sessions.Options{
		Path:     serveOptions.ContextRoot,
		MaxAge:   serveOptions.SessionMaxAge,
		HttpOnly: serveOptions.SessionHttpOnly,
		Domain:   serveOptions.SessionDomain,
		Secure:   serveOptions.SessionSecureCookies,
		SameSite: serveOptions.SessionSameSite,
	}
	log.Trace().
		Str("path", serveOptions.ContextRoot).
		Int("max_age", serveOptions.SessionMaxAge).
		Msg("Session store ready")

	otelShutdown, err := serve.SetupOTelSDK(initialContext, otelOptions)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = errors.Join(err, otelShutdown(initialContext))
	}()
	log.Trace().
		Msg("Opentelemetry ready")

	serverContext := app_http.InjectServeOptions(initialContext, serveOptions)
	oidOptionsContext := app_http.InjectOidcOptions(serverContext, oidcOptions)
	sessionStoreContext := app_http.InjectSessionStore(oidOptionsContext, sessionStore)
	var oidcContext context.Context
	log.Trace().
		Msg("Application contextes ready")

	if oidcOptions.Disabled {
		oidcContext = sessionStoreContext
		log.Trace().
			Msg("OIDC disabled")
	} else {
		relyingParty, err := serve.SetupOIDC(serveOptions, oidcOptions)
		if err != nil {
			panic(err)
		}
		log.Trace().
			Str("client_id", oidcOptions.ClientId).
			Msg("Relying party ready")

		rpContext := app_http.InjectRelyingParty(sessionStoreContext, relyingParty)
		resourceServer, err := rs.NewResourceServerClientCredentials(rpContext, oidcOptions.Issuer, oidcOptions.ClientId, oidcOptions.ClientSecret)
		if err != nil {
			panic(err)
		}
		oidcContext = app_http.InjectOidcResource(rpContext, resourceServer)

		log.Trace().
			Msg("Resource server client created")
	}

	dbClient, err := db.NewClient(dbOptions)
	if err != nil {
		panic(err)
	}
	finalContext := app_http.InjectDb(app_http.InjectDbOptions(oidcContext, dbOptions), dbClient)

	rootRouter := mux.NewRouter()
	example.Handler(rootRouter, finalContext, nil)
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
