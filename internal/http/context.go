package http

import (
	"context"

	"github.com/morphy76/g-fe-server/internal/options"
	"github.com/morphy76/g-fe-server/internal/serve"
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

func ExtractOwnership(ctx context.Context) serve.Ownership {
	return ctx.Value(ctx_OWNERSHIP_KEY).(serve.Ownership)
}

func InjectOwnership(ctx context.Context, ownership serve.Ownership) context.Context {
	return context.WithValue(ctx, ctx_OWNERSHIP_KEY, ownership)
}

func ExtractRelyingParty(ctx context.Context) rp.RelyingParty {
	return ctx.Value(ctx_OIDC_KEY).(rp.RelyingParty)
}

func InjectRelyingParty(ctx context.Context, relyingParty rp.RelyingParty) context.Context {
	return context.WithValue(ctx, ctx_OIDC_KEY, relyingParty)
}
