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
			session, err := sessions.GetRegistry(r).Get(sessionStore, sessionName)
			if err != nil {
				// TODO: handle session retrieval error
				next.ServeHTTP(w, r)
				return
			}

			accessToken, ok := session.Values[SessionKeyAccessToken].(string)
			if !ok || accessToken == "" {
				setSessionNotAuthenticated(session, r, w)
				next.ServeHTTP(w, r)
				return
			}

			if isTokenRecent(session) {
				next.ServeHTTP(w, r)
				return
			}

			if isAccessTokenValid(r, rs, accessToken) {
				next.ServeHTTP(w, r)
				return
			}

			refreshToken, ok := session.Values[SessionKeyRefreshToken].(string)
			if !ok || refreshToken == "" {
				setSessionNotAuthenticated(session, r, w)
				next.ServeHTTP(w, r)
				return
			}

			if renewTokensAndStore(session, r, w, rp, refreshToken, accessToken) {
				next.ServeHTTP(w, r)
				return
			}

			setSessionNotAuthenticated(session, r, w)
			next.ServeHTTP(w, r)
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
	const maxAge = 5 * time.Minute // adjust as needed
	issuedAtRaw, ok := session.Values[SessionKeyTokenIssuedAt]
	if !ok {
		return false
	}
	issuedAt, ok := issuedAtRaw.(int64)
	if !ok {
		return false
	}
	return time.Now().Unix()-issuedAt < int64(maxAge.Seconds())
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
	logger.Debug().Msg("Attempting to renew tokens using refresh token")

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

	logger.Debug().Msg("Token renewal successful and session updated")
	return true
}

func setSessionNotAuthenticated(session *sessions.Session, r *http.Request, w http.ResponseWriter) {
	session.Values[SessionKeyAuthenticated] = false
	session.Save(r, w)
}
