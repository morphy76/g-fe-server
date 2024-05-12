package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	factory "g-fe-server/internal/example"
	handlers "g-fe-server/internal/http"
	app_context "g-fe-server/internal/http/context"
	model "g-fe-server/pkg/example"
)

type EnvEntry string

const (
	EnvEntryContextRoot       EnvEntry = "CONTEXT_ROOT"
	EnvEntryStaticPath        EnvEntry = "STATIC_PATH"
	EnvEntryPort              EnvEntry = "PORT"
	EnvEntryHost              EnvEntry = "HOST"
	EnvEntryDBType            EnvEntry = "DB_TYPE"
	EnvEntryDBMongoUri        EnvEntry = "DB_MONGO_URL"
	EnvEntryDBMongoDb         EnvEntry = "DB_MONGO_NAME"
	EnvEntryDBMongoCollection EnvEntry = "DB_MONGO_COLLECTION"
)

func main() {

	zerolog.TimeFieldFormat = time.RFC3339
	debug := flag.Bool("trace", false, "sets log level to trace")
	ctxRootArg := flag.String("ctx", "", "presentation server context root")
	staticPathArg := flag.String("static", "/static", "static path of the served application")
	portArg := flag.String("port", "8080", "binding port of the presentation server")
	hostArg := flag.String("host", "0.0.0.0", "binding host of the presentation server")
	dbTypeArg := flag.Int("db", 0, "type of the database: 0: memory - 1: mongo")
	dbMongoUriArg := flag.String("db-mongo-uri", "", "mongo database uri in the form of mongodb://<user>:<pass>@<host>:<port>")
	dbMongoDbArg := flag.String("db-mongo-name", "", "mongo database name")
	dbMongoCollectionArg := flag.String("db-mongo-collection", "", "mongo collection to use")
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

	ctxRoot, found := os.LookupEnv(string(EnvEntryContextRoot))
	if !found {
		ctxRoot = *ctxRootArg
	}
	if len(ctxRoot) == 0 || strings.Contains(ctxRoot, " ") || !strings.HasPrefix(ctxRoot, "/") {
		fmt.Println("Invalid context root")
		os.Exit(1)
	}

	staticPath, found := os.LookupEnv(string(EnvEntryStaticPath))
	if !found {
		staticPath = *staticPathArg
	}
	if len(staticPath) == 0 || strings.Contains(staticPath, " ") {
		fmt.Println("Invalid static path")
		os.Exit(1)
	}

	usePort, found := os.LookupEnv(string(EnvEntryPort))
	if !found {
		usePort = *portArg
	}

	useHost, found := os.LookupEnv(string(EnvEntryHost))
	if !found {
		useHost = *hostArg
	}

	useDbType, found := os.LookupEnv(string(EnvEntryDBType))
	if !found {
		useDbType = fmt.Sprintf("%d", *dbTypeArg)
	}

	useDbMongoUri, found := os.LookupEnv(string(EnvEntryDBMongoUri))
	if !found {
		useDbMongoUri = *dbMongoUriArg
	}

	useDbMongo, found := os.LookupEnv(string(EnvEntryDBMongoDb))
	if !found {
		useDbMongo = *dbMongoDbArg
	}

	useDbMongoCollection, found := os.LookupEnv(string(EnvEntryDBMongoCollection))
	if !found {
		useDbMongoCollection = *dbMongoCollectionArg
	}

	useDbTypeInt, err := strconv.Atoi(useDbType)
	if err != nil {
		fmt.Println("Invalid database type")
		os.Exit(1)
	}
	dbModel := app_context.DbModel{
		Type:       model.RepositoryType(useDbTypeInt),
		Uri:        useDbMongoUri,
		Db:         useDbMongo,
		Collection: useDbMongoCollection,
	}

	ctxModel := app_context.ContextModel{
		ContextRoot: ctxRoot,
		StaticPath:  staticPath,
	}

	startServer(ctxModel, dbModel, useHost, usePort, ctxRoot, staticPath)
}

func startServer(
	ctxModel app_context.ContextModel,
	dbModel app_context.DbModel,
	useHost string,
	usePort string,
	ctxRoot string,
	staticPath string,
) {

	start := time.Now()

	repository, err := factory.NewRepository(dbModel)
	if err != nil {
		panic(err)
	}

	err = repository.Connect()
	if err != nil {
		panic(err)
	}
	defer repository.Disconnect()

	dbContext := context.WithValue(context.Background(), app_context.CTX_REPOSITORY_KEY, repository)
	serverContext := context.WithValue(dbContext, app_context.CTX_CONTEXT_ROOT_KEY, ctxModel)

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
	err = http.ListenAndServe(fmt.Sprintf("%s:%s", useHost, usePort), rootRouter)
	if err != nil {
		panic(err)
	}
}
