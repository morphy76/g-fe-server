package serve

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/morphy76/g-fe-server/internal/options"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	httphelper "github.com/zitadel/oidc/v3/pkg/http"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// SetupOIDC sets up the OIDC client
func SetupOIDC(
	serveOptions *options.ServeOptions,
	oidcOptions *options.OIDCOptions,
) (rp.RelyingParty, error) {

	redirectURI := fmt.Sprintf(
		"%s://%s:%s/%s/auth/callback",
		serveOptions.Protocol,
		serveOptions.Host,
		serveOptions.Port,
		serveOptions.ContextRoot,
	)

	cookieHandlerOpts := []httphelper.CookieHandlerOpt{
		httphelper.WithDomain(serveOptions.SessionDomain),
		httphelper.WithMaxAge(serveOptions.SessionMaxAge),
		httphelper.WithSameSite(serveOptions.SessionSameSite),
	}
	if !serveOptions.SessionSecureCookies {
		cookieHandlerOpts = append(cookieHandlerOpts, httphelper.WithUnsecure())
	}

	cookieHandler := httphelper.NewCookieHandler(
		[]byte(serveOptions.SessionKey),
		[]byte(serveOptions.SessionKey),
		cookieHandlerOpts...,
	)

	oidcOpts := []rp.Option{
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
		rp.WithHTTPClient(instrumentNewHTTPClient()),
	}

	oidcClient, err := rp.NewRelyingPartyOIDC(
		context.Background(),
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

	return oidcClient, err
}

func instrumentNewHTTPClient() *http.Client {
	transport := otelhttp.NewTransport(http.DefaultTransport)
	client := &http.Client{
		Transport: transport,
	}
	return client
}
