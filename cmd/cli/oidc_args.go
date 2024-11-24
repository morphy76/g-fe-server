package cli

import (
	"errors"
	"flag"
	"os"
	"strings"

	"github.com/morphy76/g-fe-server/internal/options"
)

// OIDCOptionsBuidlerFn is a function that returns OIDC options
type OIDCOptionsBuidlerFn func() (*options.OIDCOptions, error)

// ErrMissingIssuer is a missing OIDC issuer error
var ErrMissingIssuer = errors.New("OIDC issuer is required")

// ErrMissingClientID is a missing OIDC client id error
var ErrMissingClientID = errors.New("OIDC client id is required")

// ErrMissingClientSecret is a missing OIDC client secret error
var ErrMissingClientSecret = errors.New("OIDC client secret is required")

// IsMissingIssuer returns true if the error is due to a missing OIDC issuer
func IsMissingIssuer(err error) bool {
	return err == ErrMissingIssuer
}

// IsMissingClientID returns true if the error is due to a missing OIDC client id
func IsMissingClientID(err error) bool {
	return err == ErrMissingClientID
}

// IsMissingClientSecret returns true if the error is due to a missing OIDC client secret
func IsMissingClientSecret(err error) bool {
	return err == ErrMissingClientSecret
}

const (
	envOIDCIssuer       = "OIDC_ISSUER"
	envOIDCClientID     = "OIDC_CLIENT_ID"
	envOIDCClientSecret = "OIDC_CLIENT_SECRET"
	envOIDCScopes       = "OIDC_SCOPES"
)

// OIDCOptionsBuilder returns a function that can be used to build OIDC options
func OIDCOptionsBuilder() OIDCOptionsBuidlerFn {

	oidcDisabledArg := flag.Bool("oidc-disabled", false, "Disable OIDC.")
	oidcIssuerArg := flag.String("oidc-issuer", " ", "OIDC issuer. Environment: "+envOIDCIssuer)
	oidcClientIDArg := flag.String("oidc-client-id", " ", "OIDC client id. Environment: "+envOIDCClientID)
	oidcClientSecretArg := flag.String("oidc-client-secret", " ", "OIDC client secret. Environment: "+envOIDCClientSecret)
	oidcScopesArg := flag.String("oidc-scopes", " ", "OIDC scopes. Environment: "+envOIDCScopes)

	rv := func() (*options.OIDCOptions, error) {

		oidcDisabled := *oidcDisabledArg

		oidcIssuer, found := os.LookupEnv(envOIDCIssuer)
		if !found {
			oidcIssuer = *oidcIssuerArg
		}
		if !oidcDisabled && oidcIssuer == "" {
			return nil, ErrMissingIssuer
		}

		oidcClientID, found := os.LookupEnv(envOIDCClientID)
		if !found {
			oidcClientID = *oidcClientIDArg
		}
		if !oidcDisabled && oidcClientID == "" {
			return nil, ErrMissingClientID
		}

		oidcClientSecret, found := os.LookupEnv(envOIDCClientSecret)
		if !found {
			oidcClientSecret = *oidcClientSecretArg
		}
		if !oidcDisabled && oidcClientSecret == "" {
			return nil, ErrMissingClientSecret
		}

		oidcScopes, found := os.LookupEnv(envOIDCScopes)
		if !found {
			oidcScopes = *oidcScopesArg
		}

		return &options.OIDCOptions{
			Disabled:     oidcDisabled,
			Issuer:       oidcIssuer,
			ClientID:     oidcClientID,
			ClientSecret: oidcClientSecret,
			Scopes:       strings.Split(oidcScopes, ","),
		}, nil
	}

	return rv
}
