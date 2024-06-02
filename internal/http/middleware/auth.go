package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-jose/go-jose/v4"
	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

func InspectAndRenew(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// TODO

		session := app_http.ExtractSession(r.Context())
		relyingParty := app_http.ExtractRelyingParty(r.Context())
		logger := app_http.ExtractLogger(r.Context(), "auth")

		access_token := session.Values["access_token"]
		if access_token == nil {
			logger.Warn().Msg("No access token found")
			next.ServeHTTP(w, r)
		}
		refresh_token := session.Values["refresh_token"]
		if refresh_token == nil {
			logger.Warn().Msg("No refresh token found")
			next.ServeHTTP(w, r)
		}

		err := rp.VerifyAccessToken(access_token.(string), "", jose.ES256)
		if err == nil {
			next.ServeHTTP(w, r)
			return
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
