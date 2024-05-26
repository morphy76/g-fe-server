package handlers

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/rs/zerolog/log"

	"github.com/morphy76/g-fe-server/internal/example"
	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/morphy76/g-fe-server/internal/http/handlers/health"
	"github.com/morphy76/g-fe-server/internal/http/handlers/static"
	"github.com/morphy76/g-fe-server/internal/http/middleware"
	"github.com/morphy76/g-fe-server/internal/options"
)

func Handler(parent *mux.Router, context context.Context) {

	ctxRoot := context.Value(app_http.CTX_CONTEXT_ROOT_KEY).(*options.ServeOptions).ContextRoot
	sessionStore := context.Value(app_http.CTX_SESSION_KEY).(sessions.Store)

	middleware.InjectSession(parent, sessionStore)

	contextRouter := parent.PathPrefix(ctxRoot).Subrouter()
	if log.Trace().Enabled() {
		log.Trace().Msg("Context router registered")
	}

	contextRouter.Path("/ui").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, ctxRoot+"/ui/", http.StatusMovedPermanently)
	})
	staticRouter := contextRouter.PathPrefix("/ui/").Subrouter()
	if log.Trace().Enabled() {
		log.Trace().Msg("Static router registered")
	}

	apiRouter := contextRouter.PathPrefix("/api").Subrouter()
	if log.Trace().Enabled() {
		log.Trace().Msg("API router registered")
	}

	nonFunctionalRouter := contextRouter.PathPrefix("/g").Subrouter()
	if log.Trace().Enabled() {
		log.Trace().Msg("Non functional router registered")
	}

	apiRouter.Use(middleware.TenantResolver)
	apiRouter.Use(middleware.RequestLogger)
	apiRouter.Use(middleware.JSONResponse)
	apiRouter.Use(mux.CORSMethodMiddleware(apiRouter))
	if log.Trace().Enabled() {
		log.Trace().Msg("API middleware registered")
	}

	static.HandleStatic(staticRouter, context)
	if log.Trace().Enabled() {
		log.Trace().Msg("Static handler registered")
	}

	health.HealthHandlers(nonFunctionalRouter, context)
	if log.Trace().Enabled() {
		log.Trace().Msg("Health handler registered")
	}

	example.ExampleHandlers(apiRouter, context)
	if log.Trace().Enabled() {
		log.Trace().Msg("Example handler registered")
	}
}
