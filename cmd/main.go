package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	handlers "g-fe-server/internal/http"
	app_context "g-fe-server/internal/http/context"
)

func main() {

	start := time.Now()

	zerolog.TimeFieldFormat = time.RFC3339
	debug := flag.Bool("trace", false, "sets log level to trace")
	ctxRootArg := flag.String("ctx", "", "presentation server context root")
	staticPathArg := flag.String("static", "/static", "static path of the served application")
	portArg := flag.String("port", "8080", "binding port of the presentation server")
	hostArg := flag.String("host", "0.0.0.0", "binding host of the presentation server")
	help := flag.Bool("help", false, "prints help message")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}

	ctxRoot, found := os.LookupEnv("CONTEXT_ROOT")
	if !found {
		ctxRoot = *ctxRootArg
	}
	if len(ctxRoot) == 0 || strings.Contains(ctxRoot, " ") || !strings.HasPrefix(ctxRoot, "/") {
		fmt.Println("Invalid context root")
		os.Exit(1)
	}

	staticPath, found := os.LookupEnv("STATIC_PATH")
	if !found {
		staticPath = *staticPathArg
	}
	if len(staticPath) == 0 || strings.Contains(staticPath, " ") {
		fmt.Println("Invalid static path")
		os.Exit(1)
	}

	usePort, found := os.LookupEnv("PORT")
	if !found {
		usePort = *portArg
	}

	useHost, found := os.LookupEnv("HOST")
	if !found {
		useHost = *hostArg
	}

	ctxModel := app_context.ContextModel{
		ContextRoot: ctxRoot,
		StaticPath:  staticPath,
	}

	serverContext := context.WithValue(context.Background(), app_context.CTX_CONTEXT_ROOT_KEY, ctxModel)

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
		Str("host", useHost).
		Str("port", usePort).
		Str("ctx", ctxRoot).
		Str("serving", staticPath).
		Int64("setup_ns", time.Since(start).Nanoseconds()).
		Msg("Server started")
	err := http.ListenAndServe(fmt.Sprintf("%s:%s", useHost, usePort), rootRouter)
	if err != nil {
		panic(err)
	}
}
