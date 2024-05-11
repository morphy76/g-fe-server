package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"

	handlers "g-fe-server/internal/http"
	app_context "g-fe-server/internal/http/context"
)

func main() {

	ctxRoot, found := os.LookupEnv("CONTEXT_ROOT")
	if !found && len(os.Args) > 1 {
		ctxRoot = os.Args[1]
	}
	if len(ctxRoot) == 0 || strings.Contains(ctxRoot, " ") || !strings.HasPrefix(ctxRoot, "/") {
		fmt.Println("Invalid context root")
		os.Exit(1)
	}

	staticPath, found := os.LookupEnv("STATIC_PATH")
	if !found && len(os.Args) > 2 {
		staticPath = os.Args[2]
	}
	if len(staticPath) == 0 || strings.Contains(staticPath, " ") {
		fmt.Println("Invalid static path")
		os.Exit(1)
	}

	ctxModel := app_context.ContextModel{
		ContextRoot: ctxRoot,
		StaticPath:  staticPath,
	}

	serverContext := context.WithValue(context.Background(), app_context.CTX_CONTEXT_ROOT_KEY, ctxModel)

	rootRouter := mux.NewRouter()

	handlers.Handler(rootRouter, serverContext)

	rootRouter.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		if len(route.GetName()) > 0 {
			fmt.Printf("Endpoint: %v\n", route.GetName())
		}
		return nil
	})

	err := http.ListenAndServe(":8080", rootRouter)
	if err != nil {
		panic(err)
	} else {
		fmt.Println("Server started on port 8080")
	}
}
