package auth

import (
	"net/http"
	"net/url"

	"github.com/gorilla/sessions"
	"github.com/morphy76/g-fe-server/internal/logger"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/client/rs"
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
	rp rp.RelyingParty,
	rs rs.ResourceServer,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return isAuthenticateByBearerToken(rp, rs, isAuthenticatedBySession(ctxRoot, sessionStore, sessionName, rp, next))
	}
}

// InspectAndRenew is a middleware that inspects the session and renews it if necessary.
// TODO
func InspectAndRenew() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// touch the http session for modified and idlesince
			next.ServeHTTP(w, r)
		})
	}
}

func isAuthenticatedBySession(
	ctxRoot string,
	sessionStore sessions.Store,
	sessionName string,
	relyingParty rp.RelyingParty,
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
	rp rp.RelyingParty,
	rs rs.ResourceServer,
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

// const authLogout = "/auth/logout"

// // HTTPSessionInspectAndRenew checks the session for an active token and renews it if necessary
// func HTTPSessionInspectAndRenew(resourceServer rs.ResourceServer, relyingParty rp.RelyingParty, serveOpts *options.ServeOptions) func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			session := app_http.ExtractSession(r.Context())
// 			logger := logger.GetLogger(r.Context(), "auth")

// 			ctxRoot := serveOpts.ContextRoot
// 			requestedFile := filepath.Join(serveOpts.StaticPath, strings.TrimPrefix(r.URL.Path, ctxRoot+"/ui"))
// 			if strings.HasSuffix(requestedFile, ".js") {
// 				next.ServeHTTP(w, r)
// 				return
// 			}

// 			accessToken := session.Values[SessionKeyAccessToken]
// 			if accessToken == nil {
// 				logger.Warn().Msg("No access token found")
// 				http.Redirect(w, r, ctxRoot+authLogout, http.StatusTemporaryRedirect)
// 				return
// 			}
// 			refreshToken := session.Values[SessionKeyRefreshToken]
// 			if refreshToken == nil {
// 				logger.Warn().Msg("No refresh token found")
// 				http.Redirect(w, r, ctxRoot+authLogout, http.StatusTemporaryRedirect)
// 				return
// 			}

// 			resp, err := rs.Introspect[*oidc.IntrospectionResponse](context.Background(), resourceServer, accessToken.(string))
// 			if err != nil {
// 				logger.Warn().Err(err).Msg("Failed to refresh tokens")
// 				http.Redirect(w, r, ctxRoot+authLogout, http.StatusTemporaryRedirect)
// 				return
// 			}
// 			if resp.Active {
// 				logger.Debug().Msg("Token is active")
// 				next.ServeHTTP(w, r)
// 				return
// 			} else {
// 				logger.Debug().Msg("Token is not active")
// 			}

// 			tokens, err := rp.RefreshTokens[*oidc.IDTokenClaims](
// 				context.Background(),
// 				relyingParty,
// 				refreshToken.(string),
// 				accessToken.(string),
// 				"urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
// 			)
// 			if err != nil {
// 				logger.Error().Err(err).Msg("Failed to refresh tokens")
// 				http.Redirect(w, r, ctxRoot+authLogout, http.StatusTemporaryRedirect)
// 				return
// 			}

// 			session.Values[SessionKeyAccessToken] = tokens.AccessToken
// 			session.Values[SessionKeyIDToken] = tokens.IDToken
// 			session.Values[SessionKeyRefreshToken] = tokens.RefreshToken

// 			session.Save(r, w)

// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }

// // HTTPSessionAuthenticationRequired checks the session for an active token and redirects to the login page if necessary
// func HTTPSessionAuthenticationRequired(serveOpts *options.ServeOptions) func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			session := app_http.ExtractSession(r.Context())
// 			logger := logger.GetLogger(r.Context(), "auth")

// 			authURL := fmt.Sprintf(
// 				"%s://%s:%s/%s/auth/login?requested_url=%s",
// 				serveOpts.Protocol,
// 				serveOpts.Host,
// 				serveOpts.Port,
// 				serveOpts.ContextRoot,
// 				url.QueryEscape(r.URL.String()),
// 			)

// 			idToken := session.Values[SessionKeyIDToken]
// 			if idToken == nil || len(idToken.(string)) == 0 {
// 				logger.Debug().
// 					Str("requested_url", r.URL.String()).
// 					Msg("Redirecting to login")
// 				w.Header().Set("Cache-Control", "no-cache")
// 				http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
// 				return
// 			}

// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }
