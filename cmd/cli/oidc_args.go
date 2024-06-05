package cli

import (
	"flag"
	"os"
	"strings"

	"github.com/morphy76/g-fe-server/internal/options"
)

type oidcOptionsBuidler func() (*options.OidcOptions, error)

// var errRequiredOTLPUrl = errors.New("OTLP export enabled but no URL has been specified")

// func IsRequiredOTLPUrl(err error) bool {
// 	return err == errRequiredOTLPUrl
// }

const (
	ENV_OIDC_ISSUER        = "OIDC_ISSUER"
	ENV_OIDC_CLIENT_ID     = "OIDC_CLIENT_ID"
	ENV_OIDC_CLIENT_SECRET = "OIDC_CLIENT_SECRET"
	ENV_OIDC_SCOPES        = "OIDC_SCOPES"
)

func OidcOptionsBuilder() oidcOptionsBuidler {

	oidcIssuerArg := flag.String("oidc-issuer", "", "OIDC issuer. Environment: "+ENV_OIDC_ISSUER)
	oidcClientIdArg := flag.String("oidc-client-id", "", "OIDC client id. Environment: "+ENV_OIDC_CLIENT_ID)
	oidcClientSecretArg := flag.String("oidc-client-secret", "", "OIDC client secret. Environment: "+ENV_OIDC_CLIENT_SECRET)
	oidcScopesArg := flag.String("oidc-scopes", "", "OIDC scopes")

	rv := func() (*options.OidcOptions, error) {

		oidcIssuer, found := os.LookupEnv(ENV_OIDC_ISSUER)
		if !found {
			oidcIssuer = *oidcIssuerArg
		}

		oidcClientId, found := os.LookupEnv(ENV_OIDC_CLIENT_ID)
		if !found {
			oidcClientId = *oidcClientIdArg
		}

		oidcClientSecret, found := os.LookupEnv(ENV_OIDC_CLIENT_SECRET)
		if !found {
			oidcClientSecret = *oidcClientSecretArg
		}

		oidcScopes, found := os.LookupEnv(ENV_OIDC_SCOPES)
		if !found {
			oidcScopes = *oidcScopesArg
		}

		return &options.OidcOptions{
			Issuer:       oidcIssuer,
			ClientId:     oidcClientId,
			ClientSecret: oidcClientSecret,
			Scopes:       strings.Split(oidcScopes, ","),
		}, nil
	}

	return rv
}
