package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/quasoft/memstore"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/morphy76/g-fe-server/internal/cli"
	"github.com/morphy76/g-fe-server/internal/db"
	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/morphy76/g-fe-server/internal/http/handlers"
	"github.com/morphy76/g-fe-server/internal/options"
)

func main() {

	zerolog.TimeFieldFormat = time.RFC3339
	trace := flag.Bool("trace", false, "sets log level to trace")

	dbOptionsBuilder := cli.DbOptionsBuilder()
	serveOptionsBuilder := cli.ServeOptionsBuilder()

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

	dbOptions, err := dbOptionsBuilder()
	if err != nil {
		flag.Usage()
		os.Exit(1)
	}

	serveOptions, err := serveOptionsBuilder()
	if err != nil {
		flag.Usage()
		os.Exit(1)
	}

	store := memstore.NewMemStore([]byte(serveOptions.SessionKey))
	store.Options = &sessions.Options{
		Path:     serveOptions.ContextRoot,
		MaxAge:   serveOptions.SessionMaxAge,
		HttpOnly: serveOptions.SessionHttpOnly,
		Domain:   serveOptions.SessionDomain,
		Secure:   serveOptions.SessionSecureCookies,
		SameSite: serveOptions.SessionSameSite,
	}

	startServer(serveOptions, dbOptions, store)
}

func startServer(
	serveOptions *options.ServeOptions,
	dbOptions *options.DbOptions,
	sessionStore sessions.Store,
) {

	start := time.Now()

	dbClient, err := db.NewClient(dbOptions)
	if err != nil {
		panic(err)
	}

	serverContext := context.WithValue(context.Background(), app_http.CTX_CONTEXT_SERVE_KEY, serveOptions)
	sessionStoreContext := context.WithValue(serverContext, app_http.CTX_SESSION_STORE_KEY, sessionStore)
	dbOptsContext := context.WithValue(sessionStoreContext, app_http.CTX_DB_OPTIONS_KEY, dbOptions)
	dbContext := context.WithValue(dbOptsContext, app_http.CTX_DB_KEY, dbClient)

	rootRouter := mux.NewRouter()

	handlers.Handler(rootRouter, dbContext)

	if log.Trace().Enabled() {
		rootRouter.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			if len(route.GetName()) > 0 {
				log.Trace().Str("endpoint", route.GetName()).Msg("Endpoint registered")
			}
			return nil
		})
	}

	log.Info().
		Str("host", serveOptions.Host).
		Str("port", serveOptions.Port).
		Str("ctx", serveOptions.ContextRoot).
		Str("serving", serveOptions.StaticPath).
		Int64("setup_ns", time.Since(start).Nanoseconds()).
		Msg("Server started")
	err = http.ListenAndServe(fmt.Sprintf("%s:%s", serveOptions.Host, serveOptions.Port), rootRouter)
	if err != nil {
		panic(err)
	}
}
