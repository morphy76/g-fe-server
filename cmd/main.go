package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"

	"g-fe-server/internal/http/handlers"
)

func main() {

	if len(os.Args) == 1 {
		fmt.Println("Usage: main <context root>")
		os.Exit(1)
	}

	ctxRoot := os.Args[1]
	if strings.Contains(ctxRoot, " ") || !strings.HasPrefix(ctxRoot, "/") {
		fmt.Println("Invalid context root")
		os.Exit(1)
	}

	serverContext := context.WithValue(context.Background(), handlers.CTX_CONTEXT_ROOT_KEY, ctxRoot)

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
