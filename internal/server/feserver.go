package server

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/morphy76/g-fe-server/internal/common"
	"github.com/morphy76/g-fe-server/internal/logger"
	"github.com/morphy76/g-fe-server/internal/options"
	"github.com/morphy76/g-fe-server/internal/serve"
	"github.com/rs/zerolog"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/client/rs"

	"github.com/google/uuid"
)

const appModelCtxKey common.CtxKey = "App"

// FEServer is a simple struct that represents an event bus
type FEServer struct {
	UID string

	ServeOpts    *options.ServeOptions
	SessionStore sessions.Store

	RelayingParty  rp.RelyingParty
	ResourceServer rs.ResourceServer
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
	oidcOptions *options.OIDCOptions,
	// otelOptions *options.OtelOptions,
) context.Context {
	// shutdown, err := cli.SetupOTEL(initialContext, otelOptions)
	// defer shutdown()
	// if err != nil {
	// 	panic(err)
	// }

	feServer := &FEServer{
		UID: uuid.New().String(),

		ServeOpts:    serveOpts,
		SessionStore: sessionStore,
	}

	if !oidcOptions.Disabled {
		rp, err := serve.SetupOIDC(serveOpts, oidcOptions)
		if err != nil {
			panic(err)
		}
		feServer.RelayingParty = rp

		rs, err := rs.NewResourceServerClientCredentials(context.Background(), oidcOptions.Issuer, oidcOptions.ClientID, oidcOptions.ClientSecret)
		if err != nil {
			panic(err)
		}
		feServer.ResourceServer = rs
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

	return http.ListenAndServe(feServer.ServeOpts.Host+":"+feServer.ServeOpts.Port, rootRouter)
}

// IsOIDCEnabled returns true if OIDC is enabled
func (feServer *FEServer) IsOIDCEnabled() bool {
	return feServer.RelayingParty != nil
}

// func addMonitoring(ctx context.Context, builder zerolog.Context) zerolog.Context {
// 	feServer := server.ExtractFEServer(ctx)
// 	return builder.Dict("monitoring", zerolog.Dict().
// 		Str("fe_server_id", feServer.UID),
// 	)
// }
