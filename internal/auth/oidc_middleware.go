package auth

import (
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/morphy76/g-fe-server/internal/http/session"
	"github.com/morphy76/g-fe-server/internal/logger"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/client/rs"
)

func isAuthenticatedBySession(
	sessionName string,
	secureCookie securecookie.SecureCookie,
	sessionStore sessions.Store,
	sessionOptions *session.SessionOptions,
	relyingParty rp.RelyingParty,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		useLogger := logger.GetLogger(r.Context(), "auth")

		requestedURLStateFn := func() string {
			return r.URL.String()
		}

		useLogger.Debug().Msg("Checking session")
		cookie, err := r.Cookie(sessionName)
		if err != nil {
			rp.AuthURLHandler(requestedURLStateFn, relyingParty)(w, r)
			return
		}

		useLogger.Debug().Msg("Checking cookie")
		value := make(map[string]string)
		err = secureCookie.Decode(sessionName, cookie.Value, &value)
		if err != nil {
			rp.AuthURLHandler(requestedURLStateFn, relyingParty)(w, r)
			return
		}

		useLogger.Debug().Msg("Checking session name")
		if value["session_name"] == "" {
			expireTheCookie(w, sessionOptions)
			rp.AuthURLHandler(requestedURLStateFn, relyingParty)(w, r)
			return
		}

		useLogger.Debug().Msg("Checking session store for session " + value["session_name"])
		session, err := sessionStore.Get(r, value["session_name"])
		if err != nil {
			useLogger.Debug().Msg("Session not found with name " + value["session_name"])
			expireTheCookie(w, sessionOptions)
			rp.AuthURLHandler(requestedURLStateFn, relyingParty)(w, r)
			return
		}

		flashes := session.Flashes()
		for _, flash := range flashes {
			useLogger.Debug().Interface("flash", flash).Msg("Flash message")
		}

		useLogger.Debug().Any("values", session.Values).Msg("Checking session values")
		isAuth, found := session.Values["authenticated"]
		useLogger.Debug().Bool("found", found).Interface("auth", isAuth).Msg("Session values found")
		if found && isAuth.(string) == "true" {
			useLogger.Debug().Msg("Session is authenticated")
			renewTheCookie(w, sessionOptions, cookie.Value)
			next.ServeHTTP(w, r)
		} else {
			useLogger.Debug().Msg("Session is not authenticated")
			expireTheCookie(w, sessionOptions)
			rp.AuthURLHandler(requestedURLStateFn, relyingParty)(w, r)
		}
	})
}

func renewTheCookie(w http.ResponseWriter, sessionOptions *session.SessionOptions, value string) {
	cookie := &http.Cookie{
		Name:     sessionOptions.Name,
		Value:    value,
		Path:     sessionOptions.Path,
		MaxAge:   sessionOptions.MaxAge,
		HttpOnly: sessionOptions.HttpOnly,
		Domain:   sessionOptions.Domain,
		Secure:   sessionOptions.SecureCookies,
		SameSite: sessionOptions.SameSite,
	}
	http.SetCookie(w, cookie)
}
func expireTheCookie(w http.ResponseWriter, sessionOptions *session.SessionOptions) {
	expiredCookie := &http.Cookie{
		Name:     sessionOptions.Name,
		Value:    "",
		Path:     sessionOptions.Path,
		Domain:   sessionOptions.Domain,
		Secure:   sessionOptions.SecureCookies,
		HttpOnly: sessionOptions.HttpOnly,
		SameSite: sessionOptions.SameSite,
		MaxAge:   -1,
	}
	http.SetCookie(w, expiredCookie)
}

func isAuthenticateByBearerToken(
	rp rp.RelyingParty,
	rs rs.ResourceServer,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// token := r.Header.Get("Authorization")
		// if token == "" {
		// 	// http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		// 	return
		// }

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

func IsAuthenticated(
	rp rp.RelyingParty,
	rs rs.ResourceServer,
	sessionName string,
	secureCookie securecookie.SecureCookie,
	sessionStore sessions.Store,
	sessionOptions *session.SessionOptions,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return isAuthenticateByBearerToken(rp, rs, isAuthenticatedBySession(sessionName, secureCookie, sessionStore, sessionOptions, rp, next))
	}
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

// 			accessToken := session.Values["access_token"]
// 			if accessToken == nil {
// 				logger.Warn().Msg("No access token found")
// 				http.Redirect(w, r, ctxRoot+authLogout, http.StatusTemporaryRedirect)
// 				return
// 			}
// 			refreshToken := session.Values["refresh_token"]
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

// 			session.Values["access_token"] = tokens.AccessToken
// 			session.Values["id_token"] = tokens.IDToken
// 			session.Values["refresh_token"] = tokens.RefreshToken

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

// 			idToken := session.Values["id_token"]
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
