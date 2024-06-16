package cli

import (
	"errors"
	"flag"
	"os"
	"strings"

	"github.com/morphy76/g-fe-server/internal/options"
)

type oidcOptionsBuidler func() (*options.OidcOptions, error)

var errMissingIssuer = errors.New("OIDC issuer is required")
var errMissingClientId = errors.New("OIDC client id is required")
var errMissingClientSecret = errors.New("OIDC client secret is required")

func IsMissingIssuer(err error) bool {
	return err == errMissingIssuer
}

func IsMissingClientId(err error) bool {
	return err == errMissingClientId
}

func IsMissingClientSecret(err error) bool {
	return err == errMissingClientSecret
}

const (
	ENV_OIDC_ISSUER        = "OIDC_ISSUER"
	ENV_OIDC_CLIENT_ID     = "OIDC_CLIENT_ID"
	ENV_OIDC_CLIENT_SECRET = "OIDC_CLIENT_SECRET"
	ENV_OIDC_SCOPES        = "OIDC_SCOPES"
)

func OidcOptionsBuilder() oidcOptionsBuidler {

	oidcDisabledArg := flag.Bool("oidc-disabled", false, "Disable OIDC.")
	oidcIssuerArg := flag.String("oidc-issuer", " ", "OIDC issuer. Environment: "+ENV_OIDC_ISSUER)
	oidcClientIdArg := flag.String("oidc-client-id", " ", "OIDC client id. Environment: "+ENV_OIDC_CLIENT_ID)
	oidcClientSecretArg := flag.String("oidc-client-secret", " ", "OIDC client secret. Environment: "+ENV_OIDC_CLIENT_SECRET)
	oidcScopesArg := flag.String("oidc-scopes", " ", "OIDC scopes. Environment: "+ENV_OIDC_SCOPES)

	rv := func() (*options.OidcOptions, error) {

		oidcDisabled := *oidcDisabledArg

		oidcIssuer, found := os.LookupEnv(ENV_OIDC_ISSUER)
		if !found {
			oidcIssuer = *oidcIssuerArg
		}
		if !oidcDisabled && oidcIssuer == "" {
			return nil, errMissingIssuer
		}

		oidcClientId, found := os.LookupEnv(ENV_OIDC_CLIENT_ID)
		if !found {
			oidcClientId = *oidcClientIdArg
		}
		if !oidcDisabled && oidcClientId == "" {
			return nil, errMissingClientId
		}

		oidcClientSecret, found := os.LookupEnv(ENV_OIDC_CLIENT_SECRET)
		if !found {
			oidcClientSecret = *oidcClientSecretArg
		}
		if !oidcDisabled && oidcClientSecret == "" {
			return nil, errMissingClientSecret
		}

		oidcScopes, found := os.LookupEnv(ENV_OIDC_SCOPES)
		if !found {
			oidcScopes = *oidcScopesArg
		}

		return &options.OidcOptions{
			Disabled:     oidcDisabled,
			Issuer:       oidcIssuer,
			ClientId:     oidcClientId,
			ClientSecret: oidcClientSecret,
			Scopes:       strings.Split(oidcScopes, ","),
		}, nil
	}

	return rv
}
