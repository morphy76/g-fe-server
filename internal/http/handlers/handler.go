package handlers

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"

	"g-fe-server/internal/http/middleware"
)

type ContxtModelKey string

type ContextModel struct {
	ContextRoot string
	StaticPath  string
}

const CTX_CONTEXT_ROOT_KEY ContxtModelKey = "contextModel"

func Handler(parent *mux.Router, context context.Context) {

	ctxRoot := context.Value(CTX_CONTEXT_ROOT_KEY).(ContextModel).ContextRoot

	contextRouter := parent.PathPrefix(ctxRoot).Subrouter()

	contextRouter.Path("/ui").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, ctxRoot+"/ui/", http.StatusMovedPermanently)
	})
	staticRouter := contextRouter.PathPrefix("/ui/").Subrouter()

	apiRouter := contextRouter.PathPrefix("/api").Subrouter()

	nonFunctionalRouter := contextRouter.PathPrefix("/g").Subrouter()

	apiRouter.Use(middleware.JSONResponse)
	apiRouter.Use(mux.CORSMethodMiddleware(apiRouter))

	HandleStatic(staticRouter, context)
	HealthHandlers(nonFunctionalRouter, context)
}
