package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/morphy76/g-fe-server/internal/cli"
	"github.com/morphy76/g-fe-server/internal/example"
	handlers "github.com/morphy76/g-fe-server/internal/http"
	app_context "github.com/morphy76/g-fe-server/internal/http/context"
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
		panic(err)
	}

	serveOptions, err := serveOptionsBuilder()
	if err != nil {
		panic(err)
	}

	startServer(serveOptions, dbOptions)
}

func startServer(
	serveOptions app_context.ServeOptions,
	dbOptions app_context.DbOptions,
) {

	start := time.Now()

	repository, err := example.NewRepository(dbOptions)
	if err != nil {
		panic(err)
	}

	err = repository.Connect()
	if err != nil {
		panic(err)
	}
	defer repository.Disconnect()

	dbContext := context.WithValue(context.Background(), app_context.CTX_REPOSITORY_KEY, repository)
	serverContext := context.WithValue(dbContext, app_context.CTX_CONTEXT_ROOT_KEY, serveOptions)

	rootRouter := mux.NewRouter()

	handlers.Handler(rootRouter, serverContext)

	if log.Trace().Enabled() {
		rootRouter.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			if len(route.GetName()) > 0 {
				log.Trace().Str("endpoint", route.GetName()).Msg("Endpoint registered")
			}
			return nil
		})
	}

	log.Debug().
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
