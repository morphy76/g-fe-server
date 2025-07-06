package auth

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/sessions"
	"github.com/morphy76/g-fe-server/internal/logger"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/client/rs"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

const (
	// AuthQueryArgsRedirectTo is the query argument used to redirect after login
	AuthQueryArgsRedirectTo = "redirect_to"
)

// IsAuthenticated checks if the user is authenticated by session and bearer token.
func IsAuthenticated(
	ctxRoot string,
	sessionStore sessions.Store,
	sessionName string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return isAuthenticateByBearerToken(isAuthenticatedBySession(ctxRoot, sessionStore, sessionName, next))
	}
}

// InspectAndRenew is a middleware that inspects the session and renews it if necessary.
func InspectAndRenew(
	sessionStore sessions.Store,
	sessionName string,
	rp rp.RelyingParty,
	rs rs.ResourceServer,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			useLogger := logger.GetLogger(r.Context(), "auth")

			session, err := sessions.GetRegistry(r).Get(sessionStore, sessionName)
			if err != nil {
				useLogger.Error().Err(err).Msg("Failed to get session in InspectAndRenew")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			accessToken, ok := session.Values[SessionKeyAccessToken].(string)
			if !ok || accessToken == "" {
				useLogger.Error().Msg("Access token not found in session downstream authentication check")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if isTokenRecent(session) {
				useLogger.Trace().Msg("Token is recent, proceeding with request")
				next.ServeHTTP(w, r)
				return
			}

			if isAccessTokenValid(r, rs, accessToken) {
				useLogger.Trace().Msg("Access token is valid, proceeding with request")
				next.ServeHTTP(w, r)
				return
			}

			refreshToken, ok := session.Values[SessionKeyRefreshToken].(string)
			if !ok || refreshToken == "" {
				useLogger.Error().Msg("Refresh token not found in session, cannot renew tokens")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if renewTokensAndStore(session, r, w, rp, refreshToken, accessToken) {
				next.ServeHTTP(w, r)
				return
			}

			useLogger.Error().Msg("Failed to renew tokens, redirecting to login")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

func isAuthenticatedBySession(
	ctxRoot string,
	sessionStore sessions.Store,
	sessionName string,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		useLogger := logger.GetLogger(r.Context(), "auth")

		useSession, err := sessions.GetRegistry(r).Get(sessionStore, sessionName)
		if err != nil {
			useLogger.Error().Err(err).Msg("Start session failed")
			http.Error(w, "Start session failed", http.StatusInternalServerError)
			return
		}

		isAuth, found := useSession.Values[SessionKeyAuthenticated]
		if found && isAuth.(bool) {
			next.ServeHTTP(w, r)
		} else {
			useLogger.Debug().Msg("Session is not authenticated")
			http.Redirect(w, r,
				ctxRoot+"/auth/login?"+AuthQueryArgsRedirectTo+"="+url.QueryEscape(r.URL.String()),
				http.StatusTemporaryRedirect,
			)
		}
	})
}

func isAuthenticateByBearerToken(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		// resp, err := rs.Introspect[*oidc.IntrospectionResponse](context.Background(), resourceServer, accessToken.(string))
		// if err != nil {
		// 	logger.Warn().Err(err).Msg("Failed to refresh tokens")
		// 	http.Redirect(w, r, ctxRoot+authLogout, http.StatusTemporaryRedirect)
		// 	return
		// }
		// if resp.Active {
		// 	logger.Debug().Msg("Token is active")
		// 	next.ServeHTTP(w, r)
		// 	return
		// } else {
		// 	logger.Debug().Msg("Token is not active")
		// }

		next.ServeHTTP(w, r)
	})
}

func isTokenRecent(session *sessions.Session) bool {
	issuedAtRaw, ok := session.Values[SessionKeyTokenIssuedAt]
	if !ok {
		return false
	}
	issuedAt, ok := issuedAtRaw.(int64)
	if !ok {
		return false
	}

	expireAtRaw, ok := session.Values[SessionKeyExpiresAt]
	if !ok {
		return false
	}
	expireAt, ok := expireAtRaw.(int64)
	if !ok {
		return false
	}

	midLife := (expireAt - issuedAt) / 2
	return time.Now().Unix() <= (issuedAt + midLife)
}

func isAccessTokenValid(r *http.Request, resourceServer rs.ResourceServer, accessToken string) bool {
	ctx := r.Context()

	resp, err := rs.Introspect[*oidc.IntrospectionResponse](ctx, resourceServer, accessToken)
	if err != nil {
		return false
	}
	return resp.Active
}

func renewTokensAndStore(session *sessions.Session, r *http.Request, w http.ResponseWriter, relyingParty rp.RelyingParty, refreshToken, accessToken string) bool {
	logger := logger.GetLogger(r.Context(), "auth")
	logger.Trace().Msg("Attempting to renew tokens using refresh token")

	ctx := r.Context()
	tokens, err := rp.RefreshTokens[*oidc.IDTokenClaims](ctx, relyingParty, refreshToken, "urn:ietf:params:oauth:client-assertion-type:jwt-bearer", "")
	if err != nil {
		logger.Debug().Err(err).Msg("Token renewal failed")
		return false
	}

	session.Values[SessionKeyAccessToken] = tokens.AccessToken
	session.Values[SessionKeyIDToken] = tokens.IDToken
	session.Values[SessionKeyRefreshToken] = tokens.RefreshToken
	session.Values[SessionKeyTokenIssuedAt] = time.Now()
	session.Values[SessionKeyAuthenticated] = true
	session.Save(r, w)

	logger.Trace().Msg("Token renewal successful and session updated")
	return true
}

func setSessionNotAuthenticated(session *sessions.Session, r *http.Request, w http.ResponseWriter) {
	session.Values[SessionKeyAuthenticated] = false
	session.Save(r, w)
}
