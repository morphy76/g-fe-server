package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/client/rs"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

func InspectAndRenew(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		session := app_http.ExtractSession(r.Context())
		relyingParty := app_http.ExtractRelyingParty(r.Context())
		logger := app_http.ExtractLogger(r.Context(), "auth")
		resourceServer := app_http.ExtractOidcResource(r.Context())

		access_token := session.Values["access_token"]
		if access_token == nil {
			logger.Error().Msg("No access token found")
			http.Error(w, "No access token found", http.StatusInternalServerError)
			return
		}
		refresh_token := session.Values["refresh_token"]
		if refresh_token == nil {
			logger.Error().Msg("No refresh token found")
			http.Error(w, "No refresh token found", http.StatusInternalServerError)
			return
		}

		// TODO introspect just checks the token regardless the IDP session

		resp, err := rs.Introspect[*oidc.IntrospectionResponse](context.Background(), resourceServer, access_token.(string))
		if err != nil {
			logger.Error().Err(err).Msg("Failed to refresh tokens")
			http.Error(w, "Failed to refresh tokens", http.StatusInternalServerError)
			return
		}
		if resp.Active {
			logger.Trace().Msg("Token is active")
			next.ServeHTTP(w, r)
			return
		} else {
			logger.Trace().Msg("Token is not active")
		}

		tokens, err := rp.RefreshTokens[*oidc.IDTokenClaims](
			context.Background(),
			relyingParty,
			refresh_token.(string),
			access_token.(string),
			"urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
		)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to refresh tokens")
			http.Error(w, "Failed to refresh tokens", http.StatusInternalServerError)
			return
		}

		session.Values["access_token"] = tokens.AccessToken
		session.Values["id_token"] = tokens.IDToken
		session.Values["refresh_token"] = tokens.RefreshToken

		session.Save(r, w)

		next.ServeHTTP(w, r)
	})
}

func AuthenticationRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveOptions := app_http.ExtractServeOptions(r.Context())
		session := app_http.ExtractSession(r.Context())
		logger := app_http.ExtractLogger(r.Context(), "auth")

		authURL := fmt.Sprintf(
			"%s://%s:%s/%s/auth/login?requested_url=%s",
			serveOptions.Protocol,
			serveOptions.Host,
			serveOptions.Port,
			serveOptions.ContextRoot,
			url.QueryEscape(r.URL.String()),
		)

		idToken := session.Values["id_token"]
		if idToken == nil || len(idToken.(string)) == 0 {
			logger.Trace().
				Str("requested_url", r.URL.String()).
				Msg("Redirecting to login")
			w.Header().Set("Cache-Control", "no-cache")
			http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
			return
		}

		next.ServeHTTP(w, r)
	})
}
