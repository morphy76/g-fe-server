package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/morphy76/g-fe-server/cmd/options"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// SetupOIDC sets up the OIDC client
func SetupOIDC(
	serveOptions *options.ServeOptions,
	oidcOptions *OIDCOptions,
) (rp.RelyingParty, error) {

	ctx := context.Background()

	redirectURI := fmt.Sprintf(
		"%s://%s:%s/%s/auth/callback",
		serveOptions.Protocol,
		serveOptions.Host,
		serveOptions.Port,
		serveOptions.ContextRoot,
	)

	oidcOpts := []rp.Option{
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
		rp.WithHTTPClient(instrumentNewHTTPClient(ctx)),
	}

	relyingParty, err := rp.NewRelyingPartyOIDC(
		ctx,
		oidcOptions.Issuer,
		oidcOptions.ClientID,
		oidcOptions.ClientSecret,
		redirectURI,
		oidcOptions.Scopes,
		oidcOpts...,
	)
	if err != nil {
		return nil, err
	}

	return relyingParty, err
}

func instrumentNewHTTPClient(_ context.Context) *http.Client {
	transport := otelhttp.NewTransport(http.DefaultTransport)
	client := &http.Client{
		Transport: transport,
	}
	return client
}
