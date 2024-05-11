package handlers

import (
	"context"

	"github.com/gorilla/mux"

	"g-fe-server/internal/http/middleware"
)

type CTX_CONTEXT_ROOT string

const (
	CTX_CONTEXT_ROOT_KEY CTX_CONTEXT_ROOT = "ctxRoot"
)

func Handler(parent *mux.Router, context context.Context) {

	ctxRoot := context.Value(CTX_CONTEXT_ROOT_KEY).(string)

	contextRouter := parent.PathPrefix(ctxRoot).Subrouter()
	apiRouter := contextRouter.PathPrefix("/api").Subrouter()
	nonFunctionalRouter := contextRouter.PathPrefix("/g").Subrouter()

	apiRouter.Use(middleware.JSONResponse)
	apiRouter.Use(mux.CORSMethodMiddleware(apiRouter))

	HealthHandlers(nonFunctionalRouter, context)
}
