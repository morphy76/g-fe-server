package cli

import (
	"context"
	"errors"

	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/morphy76/g-fe-server/internal/options"
	"github.com/morphy76/g-fe-server/internal/serve"
	"github.com/rs/zerolog/log"
	"github.com/zitadel/oidc/v3/pkg/client/rs"
)

func CreateTheOIDCContext(parentContext context.Context, oidcOptions *options.OidcOptions, serveOptions *options.ServeOptions) context.Context {
	var returnContext context.Context
	if oidcOptions.Disabled {
		returnContext = parentContext
		log.Trace().
			Msg("OIDC disabled")
	} else {
		relyingParty, err := serve.SetupOIDC(serveOptions, oidcOptions)
		if err != nil {
			panic(err)
		}
		log.Trace().
			Str("client_id", oidcOptions.ClientId).
			Msg("Relying party ready")

		oidcContext := app_http.InjectRelyingParty(parentContext, relyingParty)
		resourceServer, err := rs.NewResourceServerClientCredentials(oidcContext, oidcOptions.Issuer, oidcOptions.ClientId, oidcOptions.ClientSecret)
		if err != nil {
			panic(err)
		}
		returnContext = app_http.InjectOidcResource(oidcContext, resourceServer)

		log.Trace().
			Msg("Resource server client created")
	}
	return returnContext
}

func SetupOTEL(parentContext context.Context, otelOptions *options.OtelOptions) (func(), error) {
	otelShutdown, err := serve.SetupOTelSDK(parentContext, otelOptions)
	shutdownFn := func() {
		err = errors.Join(err, otelShutdown(parentContext))
	}
	log.Trace().
		Msg("Opentelemetry ready")
	return shutdownFn, err
}
