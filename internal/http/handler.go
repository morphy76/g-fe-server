package handlers

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"

	"g-fe-server/internal/example"
	app_context "g-fe-server/internal/http/context"
	"g-fe-server/internal/http/health"
	"g-fe-server/internal/http/middleware"
	"g-fe-server/internal/http/static"
)

func Handler(parent *mux.Router, context context.Context) {

	ctxRoot := context.Value(app_context.CTX_CONTEXT_ROOT_KEY).(app_context.ContextModel).ContextRoot

	contextRouter := parent.PathPrefix(ctxRoot).Subrouter()

	contextRouter.Path("/ui").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, ctxRoot+"/ui/", http.StatusMovedPermanently)
	})
	staticRouter := contextRouter.PathPrefix("/ui/").Subrouter()

	apiRouter := contextRouter.PathPrefix("/api").Subrouter()

	nonFunctionalRouter := contextRouter.PathPrefix("/g").Subrouter()

	apiRouter.Use(middleware.TenantResolver)
	apiRouter.Use(middleware.RequestLogger)
	apiRouter.Use(middleware.JSONResponse)
	apiRouter.Use(mux.CORSMethodMiddleware(apiRouter))

	static.HandleStatic(staticRouter, context)
	health.HealthHandlers(nonFunctionalRouter, context)
	example.ExampleHandlers(apiRouter, context)
}
