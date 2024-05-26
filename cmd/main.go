package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/globalsign/mgo"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/kidstuff/mongostore"
	"github.com/quasoft/memstore"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/morphy76/g-fe-server/internal/cli"
	"github.com/morphy76/g-fe-server/internal/example"
	handlers "github.com/morphy76/g-fe-server/internal/http"
	app_context "github.com/morphy76/g-fe-server/internal/http/context"
	model "github.com/morphy76/g-fe-server/pkg/example"
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

	var store sessions.Store
	if dbOptions.Type == model.RepositoryTypeMemoryDB {
		store = memstore.NewMemStore([]byte(serveOptions.SessionKey))
	} else {
		useUrl, err := url.Parse(dbOptions.Url)
		if err != nil {
			panic(err)
		}
		if useUrl.User == nil {
			useCredentials := url.UserPassword(dbOptions.User, dbOptions.Password)
			useUrl.User = useCredentials
		}

		fmt.Printf("Connecting to %s\n", useUrl.String())
		dbsess, err := mgo.DialWithInfo(&mgo.DialInfo{
			Addrs:    []string{"localhost:27017"},
			Database: "go_db",
			Username: "go",
			Password: "go",
			Timeout:  60 * time.Second,
		})

		if err != nil {
			panic(err)
		}
		defer dbsess.Close()

		useDbSess := dbsess.DB("go_db").C("sessions")
		store = mongostore.NewMongoStore(useDbSess, 3600, true, []byte(serveOptions.SessionKey))
	}

	startServer(serveOptions, dbOptions, store)
}

func startServer(
	serveOptions app_context.ServeOptions,
	dbOptions app_context.DbOptions,
	sessionStore sessions.Store,
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

	serverContext := context.WithValue(context.Background(), app_context.CTX_CONTEXT_ROOT_KEY, serveOptions)
	sessionContext := context.WithValue(serverContext, app_context.CTX_SESSION_KEY, sessionStore)
	finalContext := context.WithValue(sessionContext, app_context.CTX_REPOSITORY_KEY, repository)

	rootRouter := mux.NewRouter()

	handlers.Handler(rootRouter, finalContext)

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
