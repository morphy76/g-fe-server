package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/morphy76/g-fe-server/internal/common"
	"github.com/morphy76/g-fe-server/internal/logger"
	"github.com/morphy76/g-fe-server/internal/options"
	"github.com/rs/zerolog"

	"github.com/google/uuid"
)

const appModelCtxKey common.CtxKey = "App"

// FEServer is a simple struct that represents an event bus
type FEServer struct {
	UID string

	NonFunctionalRoot string

	ServeOpts    *options.ServeOptions
	SessionStore sessions.Store
}

// ExtractFEServer returns the FEServer from the context
func ExtractFEServer(ctx context.Context) *FEServer {
	return ctx.Value(appModelCtxKey).(*FEServer)
}

// NewFEServer creates a Context with a new EventBus
func NewFEServer(
	ctx context.Context,
	serveOpts *options.ServeOptions,
	sessionStore sessions.Store,
	oidcOptions *options.OidcOptions,
	otelOptions *options.OtelOptions,
) context.Context {

	// shutdown, err := cli.SetupOTEL(initialContext, otelOptions)
	// defer shutdown()
	// if err != nil {
	// 	panic(err)
	// }

	// serverContext := app_http.InjectServeOptions(initialContext, serveOptions)
	// oidcOptionsContext := app_http.InjectOidcOptions(serverContext, oidcOptions)
	// sessionStoreContext := app_http.InjectSessionStore(oidcOptionsContext, sessionStore)
	// finalContext := cli.CreateTheOIDCContext(sessionStoreContext, oidcOptions, serveOptions)
	// log.Trace().
	// 	Msg("Application contextes ready")

	feServer := &FEServer{
		UID: uuid.New().String(),

		//TODO: get it from serve options
		NonFunctionalRoot: "/g",

		ServeOpts:    serveOpts,
		SessionStore: sessionStore,
	}

	return context.WithValue(ctx, appModelCtxKey, feServer)
}

// ListenAndServe starts the server
func (feServer *FEServer) ListenAndServe(ctx context.Context, rootRouter *mux.Router) error {

	feLogger := logger.GetLogger(ctx, "feServer")
	feLogger.Info().Dict("serve_opts", zerolog.Dict().
		Str("host", feServer.ServeOpts.Host).
		Str("port", feServer.ServeOpts.Port).
		Str("ctx", feServer.ServeOpts.ContextRoot).
		Str("serving", feServer.ServeOpts.StaticPath)).
		Msg("Server started")

	return http.ListenAndServe(fmt.Sprintf("%s:%s", feServer.ServeOpts.Host, feServer.ServeOpts.Port), rootRouter)
}

// func addMonitoring(ctx context.Context, builder zerolog.Context) zerolog.Context {
// 	feServer := server.ExtractFEServer(ctx)
// 	return builder.Dict("monitoring", zerolog.Dict().
// 		Str("fe_server_id", feServer.UID),
// 	)
// }
