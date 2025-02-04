package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/morphy76/g-fe-server/cmd/options"
	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/morphy76/g-fe-server/internal/logger"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/client/rs"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

const authLogout = "/auth/logout"

// HTTPSessionInspectAndRenew checks the session for an active token and renews it if necessary
func HTTPSessionInspectAndRenew(resourceServer rs.ResourceServer, relyingParty rp.RelyingParty, serveOpts *options.ServeOptions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := app_http.ExtractSession(r.Context())
			logger := logger.GetLogger(r.Context(), "auth")

			ctxRoot := serveOpts.ContextRoot
			requestedFile := filepath.Join(serveOpts.StaticPath, strings.TrimPrefix(r.URL.Path, ctxRoot+"/ui"))
			if strings.HasSuffix(requestedFile, ".js") {
				next.ServeHTTP(w, r)
				return
			}

			accessToken := session.Values["access_token"]
			if accessToken == nil {
				logger.Warn().Msg("No access token found")
				http.Redirect(w, r, ctxRoot+authLogout, http.StatusTemporaryRedirect)
				return
			}
			refreshToken := session.Values["refresh_token"]
			if refreshToken == nil {
				logger.Warn().Msg("No refresh token found")
				http.Redirect(w, r, ctxRoot+authLogout, http.StatusTemporaryRedirect)
				return
			}

			resp, err := rs.Introspect[*oidc.IntrospectionResponse](context.Background(), resourceServer, accessToken.(string))
			if err != nil {
				logger.Warn().Err(err).Msg("Failed to refresh tokens")
				http.Redirect(w, r, ctxRoot+authLogout, http.StatusTemporaryRedirect)
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
				refreshToken.(string),
				accessToken.(string),
				"urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
			)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to refresh tokens")
				http.Redirect(w, r, ctxRoot+authLogout, http.StatusTemporaryRedirect)
				return
			}

			session.Values["access_token"] = tokens.AccessToken
			session.Values["id_token"] = tokens.IDToken
			session.Values["refresh_token"] = tokens.RefreshToken

			session.Save(r, w)

			next.ServeHTTP(w, r)
		})
	}
}

// HTTPSessionAuthenticationRequired checks the session for an active token and redirects to the login page if necessary
func HTTPSessionAuthenticationRequired(serveOpts *options.ServeOptions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := app_http.ExtractSession(r.Context())
			logger := logger.GetLogger(r.Context(), "auth")

			authURL := fmt.Sprintf(
				"%s://%s:%s/%s/auth/login?requested_url=%s",
				serveOpts.Protocol,
				serveOpts.Host,
				serveOpts.Port,
				serveOpts.ContextRoot,
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
}
