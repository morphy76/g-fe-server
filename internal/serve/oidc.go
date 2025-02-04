package serve

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/morphy76/g-fe-server/cmd/options"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	httphelper "github.com/zitadel/oidc/v3/pkg/http"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// SetupOIDC sets up the OIDC client
func SetupOIDC(
	serveOptions *options.ServeOptions,
	sessionOptions *options.SessionOptions,
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
		httphelper.WithDomain(sessionOptions.SessionDomain),
		httphelper.WithMaxAge(sessionOptions.SessionMaxAge),
		httphelper.WithSameSite(sessionOptions.SessionSameSite),
	}
	if !sessionOptions.SessionSecureCookies {
		cookieHandlerOpts = append(cookieHandlerOpts, httphelper.WithUnsecure())
	}

	cookieHandler := httphelper.NewCookieHandler(
		[]byte(sessionOptions.SessionKey),
		[]byte(sessionOptions.SessionKey),
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
