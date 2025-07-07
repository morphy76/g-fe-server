package auth

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
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

// RoleCheckType defines the type for role checking logic.
type RoleCheckType bool

// RoleCheckTypeAnd is a constant for role checking type that requires all roles to match.
var RoleCheckTypeAnd RoleCheckType = true

// RoleCheckTypeOr is a constant for role checking type that allows any role match.
var RoleCheckTypeOr RoleCheckType = false

// IsAuthenticatedFn defines a function type for authentication middleware.
type IsAuthenticatedFn func() mux.MiddlewareFunc

// InspectAndRenewFn defines a function type for inspecting and renewing sessions.
type InspectAndRenewFn func() mux.MiddlewareFunc

// UserInRolesFn defines a function type for checking if a user is in specified roles.
type UserInRolesFn func(roles []string, checkType RoleCheckType) mux.MiddlewareFunc

// HasAuthorizationByURIFn defines a function type for checking if a user has access to a UMA resource by its URI.
type HasAuthorizationByURIFn func(resourceURI string, scope string) mux.MiddlewareFunc

// HasAuthorizationByTypeFn defines a function type for checking if a user has access to a UMA resource by its type.
type HasAuthorizationByTypeFn func(resourceType string, scope string) mux.MiddlewareFunc

// OIDCMiddleWare holds the authentication middleware functions.
type OIDCMiddleWare struct {
	IsAuthenticated        IsAuthenticatedFn
	InspectAndRenew        InspectAndRenewFn
	UserInRoles            UserInRolesFn
	HasAuthorizationByURI  HasAuthorizationByURIFn
	HasAuthorizationByType HasAuthorizationByTypeFn
}

// NewOIDCMiddleWare creates a new OIDCMiddleWare instance with the provided parameters.
func NewOIDCMiddleWare(
	ctxRoot string,
	sessionStore sessions.Store,
	sessionName string,
	relyingParty rp.RelyingParty,
	resourceServer rs.ResourceServer,
) OIDCMiddleWare {
	return OIDCMiddleWare{
		IsAuthenticated: func() mux.MiddlewareFunc {
			return isAuthenticated(ctxRoot, sessionStore, sessionName)
		},
		InspectAndRenew: func() mux.MiddlewareFunc {
			return inspectAndRenew(sessionStore, sessionName, relyingParty, resourceServer)
		},
		UserInRoles: func(roles []string, checkType RoleCheckType) mux.MiddlewareFunc {
			return userInRoles(roles, checkType, sessionStore, sessionName)
		},
		HasAuthorizationByURI: func(resourceURI string, scope string) mux.MiddlewareFunc {
			return hasAuthorizationByURI(resourceURI, scope, sessionStore, sessionName, relyingParty, resourceServer)
		},
		HasAuthorizationByType: func(resourceType string, scope string) mux.MiddlewareFunc {
			return hasAuthorizationByType(resourceType, scope, sessionStore, sessionName, relyingParty, resourceServer)
		},
	}
}

func isAuthenticated(
	ctxRoot string,
	sessionStore sessions.Store,
	sessionName string,
) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return isAuthenticateByBearerToken(isAuthenticatedBySession(ctxRoot, sessionStore, sessionName, next))
	}
}

func inspectAndRenew(
	sessionStore sessions.Store,
	sessionName string,
	relyingParty rp.RelyingParty,
	resourceServer rs.ResourceServer,
) mux.MiddlewareFunc {
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

			if isAccessTokenValid(r, resourceServer, accessToken) {
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

			if renewTokensAndStore(session, r, w, relyingParty, refreshToken, accessToken) {
				next.ServeHTTP(w, r)
				return
			}

			useLogger.Error().Msg("Failed to renew tokens, redirecting to login")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}

func userInRoles(
	roles []string,
	checkType RoleCheckType,
	sessionStore sessions.Store,
	sessionName string,
) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			useLogger := logger.GetLogger(r.Context(), "auth")

			useSession, err := sessions.GetRegistry(r).Get(sessionStore, sessionName)
			if err != nil {
				useLogger.Error().Err(err).Msg("Start session failed")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			userInfo, ok := useSession.Values[SessionKeyUserInfo].(*UserInfo)
			if !ok || userInfo == nil {
				useLogger.Error().Msg("Session does not contain user info")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// if len(roles) > 0 {
			// 	userRoles := userInfo.ResourceAccess
			// 	if userRoles == nil {
			// 		http.Error(w, "Forbidden", http.StatusForbidden)
			// 		return
			// 	}

			// 	roleFound := false
			// 	for _, role := range roles {
			// 		if checkType == RoleCheckTypeAnd {
			// 			if _, exists := userRoles[role]; !exists {
			// 				http.Error(w, "Forbidden", http.StatusForbidden)
			// 				return
			// 			}
			// 		} else if checkType == RoleCheckTypeOr {
			// 			if _, exists := userRoles[role]; exists {
			// 				roleFound = true
			// 				break
			// 			}
			// 		}
			// 	}

			// 	if checkType == RoleCheckTypeOr && !roleFound {
			// 		http.Error(w, "Forbidden", http.StatusForbidden)
			// 		return
			// 	}
			// }

			next.ServeHTTP(w, r)
		})
	}
}

func hasAuthorizationByURI(
	resourceURI string,
	scope string,
	sessionStore sessions.Store,
	sessionName string,
	relyingParty rp.RelyingParty,
	resourceServer rs.ResourceServer,
) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			useLogger := logger.GetLogger(r.Context(), "auth")

			useSession, err := sessions.GetRegistry(r).Get(sessionStore, sessionName)
			if err != nil {
				useLogger.Error().Err(err).Msg("Start session failed")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			accessToken, ok := useSession.Values[SessionKeyAccessToken].(string)
			if !ok || accessToken == "" {
				useLogger.Error().Msg("Access token not found in session")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if !isAccessTokenValid(r, resourceServer, accessToken) {
				useLogger.Error().Msg("Access token is not valid")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// TODO
			// resp, err := rs.HasAuthorizationByURI(r.Context(), resourceServer, accessToken, resourceURI, scope)
			// if err != nil || !resp {
			// 	useLogger.Error().Err(err).Msg("Failed to check UMA resource by URI")
			// 	http.Error(w, "Forbidden", http.StatusForbidden)
			// 	return
			// }

			next.ServeHTTP(w, r)
		})
	}
}

func hasAuthorizationByType(
	resourceType string,
	scope string,
	sessionStore sessions.Store,
	sessionName string,
	relyingParty rp.RelyingParty,
	resourceServer rs.ResourceServer,
) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			useLogger := logger.GetLogger(r.Context(), "auth")

			useSession, err := sessions.GetRegistry(r).Get(sessionStore, sessionName)
			if err != nil {
				useLogger.Error().Err(err).Msg("Start session failed")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			accessToken, ok := useSession.Values[SessionKeyAccessToken].(string)
			if !ok || accessToken == "" {
				useLogger.Error().Msg("Access token not found in session")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if !isAccessTokenValid(r, resourceServer, accessToken) {
				useLogger.Error().Msg("Access token is not valid")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// TODO
			// resp, err := rs.HasAuthorizationByType(r.Context(), resourceServer, accessToken, resourceType, scope)
			// if err != nil || !resp {
			// 	useLogger.Error().Err(err).Msg("Failed to check UMA resource by type")
			// 	http.Error(w, "Forbidden", http.StatusForbidden)
			// 	return
			// }

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

	resp, _ := rs.Introspect[*oidc.IntrospectionResponse](ctx, resourceServer, accessToken)

	return resp.Active
}

func renewTokensAndStore(session *sessions.Session, r *http.Request, w http.ResponseWriter, relyingParty rp.RelyingParty, refreshToken, accessToken string) bool {
	logger := logger.GetLogger(r.Context(), "auth")
	logger.Trace().Msg("Attempting to renew tokens using refresh token")

	ctx := r.Context()
	tokens, err := rp.RefreshTokens[*oidc.IDTokenClaims](ctx, relyingParty, refreshToken, "urn:ietf:params:oauth:client-assertion-type:jwt-bearer", "")
	if err != nil {
		logger.Warn().Err(err).Msg("Token renewal failed")
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
