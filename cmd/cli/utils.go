package cli

import (
	"context"
	"errors"

	"github.com/gorilla/sessions"
	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/morphy76/g-fe-server/internal/options"
	"github.com/morphy76/g-fe-server/internal/serve"
	"github.com/quasoft/memstore"
	"github.com/rs/zerolog/log"
	"github.com/zitadel/oidc/v3/pkg/client/rs"
)

func CreateTheOIDCContext(sessionStoreContext context.Context, oidcOptions *options.OidcOptions, serveOptions *options.ServeOptions) context.Context {
	var finalContext context.Context
	if oidcOptions.Disabled {
		finalContext = sessionStoreContext
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

		oidcContext := app_http.InjectRelyingParty(sessionStoreContext, relyingParty)
		resourceServer, err := rs.NewResourceServerClientCredentials(oidcContext, oidcOptions.Issuer, oidcOptions.ClientId, oidcOptions.ClientSecret)
		if err != nil {
			panic(err)
		}
		finalContext = app_http.InjectOidcResource(oidcContext, resourceServer)

		log.Trace().
			Msg("Resource server client created")
	}
	return finalContext
}

func SetupOTEL(initialContext context.Context, otelOptions *options.OtelOptions) (func(), error) {
	otelShutdown, err := serve.SetupOTelSDK(initialContext, otelOptions)
	shutdownFn := func() {
		err = errors.Join(err, otelShutdown(initialContext))
	}
	log.Trace().
		Msg("Opentelemetry ready")
	return shutdownFn, err
}

func CreateSessionStore(serveOptions *options.ServeOptions) *memstore.MemStore {
	sessionStore := memstore.NewMemStore([]byte(serveOptions.SessionKey))
	sessionStore.Options = &sessions.Options{
		Path:     serveOptions.ContextRoot,
		MaxAge:   serveOptions.SessionMaxAge,
		HttpOnly: serveOptions.SessionHttpOnly,
		Domain:   serveOptions.SessionDomain,
		Secure:   serveOptions.SessionSecureCookies,
		SameSite: serveOptions.SessionSameSite,
	}
	log.Trace().
		Str("path", serveOptions.ContextRoot).
		Int("max_age", serveOptions.SessionMaxAge).
		Msg("Session store ready")
	return sessionStore
}
