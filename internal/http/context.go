package http

import (
	"context"

	"github.com/gorilla/sessions"
	"github.com/morphy76/g-fe-server/internal/options"
	"github.com/morphy76/g-fe-server/internal/serve"
	"github.com/rs/zerolog"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/client/rs"
)

type ContextModelKey string
type ContextSessionKey string
type ContextSessionStoreKey string
type ContextLoggerKey string
type ContextOwnershipKey string
type ContextOIDCOptions string
type ContextOIDCKey string
type ContextOIDCResourceKey string
type RouteChannelKey string

const (
	ctx_CONTEXT_SERVE_KEY ContextModelKey        = "contextModel"
	ctx_SESSION_STORE_KEY ContextSessionStoreKey = "sessionStore"
	ctx_SESSION_KEY       ContextSessionKey      = "session"
	ctx_LOGGER_KEY        ContextLoggerKey       = "logger"
	ctx_OWNERSHIP_KEY     ContextOwnershipKey    = "ownership"
	ctx_OIDC_OPTIONS_KEY  ContextOIDCOptions     = "oidcOptions"
	ctx_OIDC_KEY          ContextOIDCKey         = "oidc"
	ctx_OIDC_RESOURCE_KEY ContextOIDCResourceKey = "oidcResource"
	ctx_ROUTE_CHANNEL_KEY RouteChannelKey        = "routeChannel"
)

func InjectOidcOptions(ctx context.Context, oidcOptions *options.OidcOptions) context.Context {
	return context.WithValue(ctx, ctx_OIDC_OPTIONS_KEY, oidcOptions)
}

func ExtractOidcOptions(ctx context.Context) *options.OidcOptions {
	return ctx.Value(ctx_OIDC_OPTIONS_KEY).(*options.OidcOptions)
}

func InjectOidcResource(ctx context.Context, resource rs.ResourceServer) context.Context {
	return context.WithValue(ctx, ctx_OIDC_RESOURCE_KEY, resource)
}

func ExtractOidcResource(ctx context.Context) rs.ResourceServer {
	return ctx.Value(ctx_OIDC_RESOURCE_KEY).(rs.ResourceServer)
}

func ExtractServeOptions(ctx context.Context) *options.ServeOptions {
	return ctx.Value(ctx_CONTEXT_SERVE_KEY).(*options.ServeOptions)
}

func InjectServeOptions(ctx context.Context, serveOptions *options.ServeOptions) context.Context {
	return context.WithValue(ctx, ctx_CONTEXT_SERVE_KEY, serveOptions)
}

func ExtractLogger(ctx context.Context, forPackage string) zerolog.Logger {
	return (ctx.Value(ctx_LOGGER_KEY).(zerolog.Logger)).With().Str("package", forPackage).Logger()
}

func InjectLogger(ctx context.Context, logger zerolog.Logger) context.Context {
	return context.WithValue(ctx, ctx_LOGGER_KEY, logger)
}

func ExtractOwnership(ctx context.Context) serve.Ownership {
	return ctx.Value(ctx_OWNERSHIP_KEY).(serve.Ownership)
}

func InjectOwnership(ctx context.Context, ownership serve.Ownership) context.Context {
	return context.WithValue(ctx, ctx_OWNERSHIP_KEY, ownership)
}

func ExtractSessionStore(ctx context.Context) sessions.Store {
	return ctx.Value(ctx_SESSION_STORE_KEY).(sessions.Store)
}

func InjectSessionStore(ctx context.Context, store sessions.Store) context.Context {
	return context.WithValue(ctx, ctx_SESSION_STORE_KEY, store)
}

func ExtractSession(ctx context.Context) *sessions.Session {
	return ctx.Value(ctx_SESSION_KEY).(*sessions.Session)
}

func InjectSession(ctx context.Context, session *sessions.Session) context.Context {
	return context.WithValue(ctx, ctx_SESSION_KEY, session)
}

func ExtractRelyingParty(ctx context.Context) rp.RelyingParty {
	return ctx.Value(ctx_OIDC_KEY).(rp.RelyingParty)
}

func InjectRelyingParty(ctx context.Context, relyingParty rp.RelyingParty) context.Context {
	return context.WithValue(ctx, ctx_OIDC_KEY, relyingParty)
}
