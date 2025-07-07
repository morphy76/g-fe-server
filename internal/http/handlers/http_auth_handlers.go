package handlers

import (
	"context"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/morphy76/g-fe-server/cmd/options"
	"github.com/morphy76/g-fe-server/internal/auth"
	"github.com/morphy76/g-fe-server/internal/logger"
	"github.com/morphy76/g-fe-server/internal/server"
	"github.com/rs/zerolog"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/client/rs"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

const (
	failedToGetSession        = "Failed to get session"
	internalServerError       = "Internal server error"
	logoutTokenParam          = "logout_token="
	backChannelLogoutEvent    = "http://schemas.openid.net/event/backchannel-logout"
	sessionCleanupTimeout     = 5 * time.Second
	jtiCheckTimeout           = 2 * time.Second
	jtiReplayCheckWindow      = 5 * time.Minute
	httpSessionsCollection    = "http_sessions"
	logoutTokenJtisCollection = "logout_token_jtis"
)

type sessionStateStruct struct {
	RedirectTo string `json:"redirect_to"`
	SID        string `json:"sid"`
}

func validateRedirectURL(redirectURL string, ctxRoot string) string {
	if redirectURL == "" {
		return ctxRoot + "/ui"
	}

	parsedURL, err := url.Parse(redirectURL)
	if err != nil {
		return ctxRoot + "/ui"
	}

	if parsedURL.Scheme != "" || parsedURL.Host != "" {
		return ctxRoot + "/ui"
	}

	if strings.HasPrefix(redirectURL, "/") && len(redirectURL) > 1 && (redirectURL[1] == '/' || redirectURL[1] == '\\') {
		return ctxRoot + "/ui"
	}

	if !strings.HasPrefix(redirectURL, ctxRoot) && !strings.HasPrefix(redirectURL, "/") {
		redirectURL = ctxRoot + "/" + strings.TrimPrefix(redirectURL, "/")
	}

	if strings.Contains(redirectURL, "://") || strings.Contains(redirectURL, "javascript:") || strings.Contains(redirectURL, "data:") {
		return ctxRoot + "/ui"
	}

	return redirectURL
}

// IAMHandlers registers the IAM authentication handlers
func IAMHandlers(
	authRouter *mux.Router,
	httpOptions *options.HTTPOptions,
	sessionStore sessions.Store,
	relyingParty rp.RelyingParty,
	resourceServer rs.ResourceServer,
) error {
	ctxRoot := httpOptions.ServeOptions.ContextRoot

	gob.Register(auth.UserInfo{})
	gob.Register(time.Time{})

	authRouter.HandleFunc("/login", onLogin(sessionStore, httpOptions, ctxRoot, relyingParty)).Name("GET " + ctxRoot + "/auth/login")
	authRouter.HandleFunc("/callback", rp.CodeExchangeHandler(rp.UserinfoCallback(buildUserInfoCallback(resourceServer)), relyingParty)).Name("GET " + ctxRoot + "/auth/callback")
	authRouter.HandleFunc("/logout", onLogout(sessionStore, httpOptions, ctxRoot, relyingParty)).Name("GET " + ctxRoot + "/auth/logout")
	authRouter.HandleFunc("/info", onInfo(sessionStore, httpOptions)).Name("GET " + ctxRoot + "/auth/info")
	authRouter.HandleFunc("/bc_logout", onBackChannelLogout()).Methods("POST").Name("POST " + ctxRoot + "/auth/bc_logout")

	return nil
}

func onLogin(sessionStore sessions.Store, httpOptions *options.HTTPOptions, ctxRoot string, relyingParty rp.RelyingParty) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logger.GetLogger(r.Context(), "auth")
		log.Trace().Msg("Logging in")

		requestedURL, err := url.QueryUnescape(r.URL.Query().Get(auth.AuthQueryArgsRedirectTo))
		if err != nil {
			requestedURL = ctxRoot + "/ui"
		}

		requestedURL = validateRedirectURL(requestedURL, ctxRoot)

		session, err := sessions.GetRegistry(r).Get(sessionStore, httpOptions.SessionOptions.Name)
		if err != nil {
			log.Error().Err(err).Msg("Start session failed")
			http.Error(w, "Start session failed", http.StatusInternalServerError)
			return
		}

		isAuth, found := session.Values[auth.SessionKeyAuthenticated]
		if found && isAuth.(bool) {
			log.Debug().Msg("Session is already authenticated")
			http.Redirect(w, r, requestedURL, http.StatusFound)
			return
		}

		rp.AuthURLHandler(func() string {
			state, _ := json.Marshal(sessionStateStruct{
				RedirectTo: requestedURL,
				SID:        session.ID,
			})
			// Base64 encode the state for safe transport
			return base64.URLEncoding.EncodeToString(state)
		}, relyingParty)(w, r)
	}
}

func onLogout(sessionStore sessions.Store, httpOptions *options.HTTPOptions, ctxRoot string, relyingParty rp.RelyingParty) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logger.GetLogger(r.Context(), "auth")
		log.Trace().Msg("Logging out")

		session, err := sessions.GetRegistry(r).Get(sessionStore, httpOptions.SessionOptions.Name)
		if err != nil {
			log.Error().Err(err).Msg(failedToGetSession)
			http.Error(w, failedToGetSession, http.StatusInternalServerError)
			return
		}

		issuer := session.Values[auth.SessionKeyIssuer]
		subject := session.Values[auth.SessionKeySubject]
		sid := session.Values[auth.SessionKeySessionID]
		idToken := session.Values[auth.SessionKeyIDToken]
		useSessionState := &sessionStateStruct{}
		json.Unmarshal([]byte(session.Values[auth.SessionKeySessionState].(string)), useSessionState)
		requestedURL := validateRedirectURL(useSessionState.RedirectTo, ctxRoot)

		log.Debug().
			Interface("issuer", issuer).
			Interface("subject", subject).
			Interface("session_id", sid).
			Interface("id_token", idToken).
			Str("redirect_to", requestedURL).
			Msg("Start logging out")

		encodedSessionState := base64.URLEncoding.EncodeToString([]byte(session.Values[auth.SessionKeySessionState].(string)))
		url, err := rp.EndSession(r.Context(), relyingParty, idToken.(string), requestedURL, encodedSessionState)
		if err != nil {
			log.Error().Err(err).Msg("End session failed")
			http.Error(w, "End session failed", http.StatusInternalServerError)
			return
		}

		session.Options.MaxAge = -1
		session.Save(r, w)
		log.Trace().
			Any("to url", url).
			Msg("Auth session deleted")

		http.Redirect(w, r, url.String(), http.StatusFound)
	}
}

func onInfo(sessionStore sessions.Store, httpOptions *options.HTTPOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logger.GetLogger(r.Context(), "auth")

		session, err := sessions.GetRegistry(r).Get(sessionStore, httpOptions.SessionOptions.Name)
		if err != nil {
			log.Error().Err(err).Msg(failedToGetSession)
			http.Error(w, failedToGetSession, http.StatusInternalServerError)
			return
		}

		log.Trace().Msg("Info requested")

		authenticated, found := session.Values[auth.SessionKeyAuthenticated]
		if !found || !authenticated.(bool) {
			log.Debug().Msg("User is not authenticated")
			http.Error(w, "User is not authenticated", http.StatusUnauthorized)
			return
		}

		rv := session.Values[auth.SessionKeyUserInfo]
		responseBody, err := json.Marshal(rv)
		if err != nil {
			log.Error().Err(err).Msg("Failed to marshal response")
			http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseBody)
	}
}

func onBackChannelLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logger.GetLogger(r.Context(), "auth")
		log.Debug().Msg("Back channel logout")

		w.Header().Set("Cache-Control", "no-store")

		logoutToken, err := extractLogoutToken(r, log)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		claims, err := parseLogoutToken(logoutToken, log)
		if err != nil {
			http.Error(w, "Invalid logout_token", http.StatusBadRequest)
			return
		}

		feServer, err := server.ExtractFEServer(r.Context())
		if err != nil {
			log.Error().Err(err).Msg("Failed to extract FEServer from context")
			http.Error(w, internalServerError, http.StatusInternalServerError)
			return
		}

		if err := validateLogoutToken(claims, log, feServer); err != nil {
			log.Error().Err(err).Msg("Logout token validation failed")
			http.Error(w, "Invalid logout_token", http.StatusBadRequest)
			return
		}

		log.Debug().
			Str("subject", claims.Subject).
			Str("issuer", claims.Issuer).
			Str("session_id", claims.SessionID).
			Str("jti", claims.JWTID).
			Msg("Decoded logout_token")

		go cleanupUserSessions(claims, feServer, log)

		w.WriteHeader(http.StatusOK)
	}
}

func validateLogoutToken(claims *oidc.LogoutTokenClaims, log zerolog.Logger, feServer *server.FEServer) error {
	// 1. JWT signature validation is handled by oidc.ParseToken

	// 2-4. Validate required claims and expiration
	if err := validateBasicClaims(claims); err != nil {
		return err
	}

	// 5. Verify events claim - CRITICAL for compliance
	if err := validateEventsClaim(claims); err != nil {
		return err
	}

	// 6. Verify nonce claim is not present (LogoutTokenClaims doesn't have Nonce field, so this is automatically satisfied)

	// 7. Verify JWTID is present (required for logout tokens)
	if claims.JWTID == "" {
		return fmt.Errorf("missing required 'jti' claim")
	}

	// 8. Optional: verify issuer matches expected issuer
	if err := validateIssuer(claims, feServer, log); err != nil {
		return err
	}

	// 9. Optional: verify JTI uniqueness to prevent replay attacks
	if err := validateJTIUniqueness(claims, feServer, log); err != nil {
		return err
	}

	// 10. Optional: verify subject and session ID match existing sessions
	// This could be implemented to cross-reference with your session store

	return nil
}

func validateBasicClaims(claims *oidc.LogoutTokenClaims) error {
	if claims.Issuer == "" {
		return fmt.Errorf("missing required 'iss' claim")
	}
	if len(claims.Audience) == 0 {
		return fmt.Errorf("missing required 'aud' claim")
	}
	if claims.IssuedAt == 0 {
		return fmt.Errorf("missing required 'iat' claim")
	}
	if claims.Expiration == 0 {
		return fmt.Errorf("missing required 'exp' claim")
	}

	if time.Now().After(claims.Expiration.AsTime()) {
		return fmt.Errorf("logout token has expired")
	}

	if claims.Subject == "" && claims.SessionID == "" {
		return fmt.Errorf("logout token must contain either 'sub' or 'sid' claim")
	}

	return nil
}

func validateEventsClaim(claims *oidc.LogoutTokenClaims) error {
	if claims.Events == nil {
		return fmt.Errorf("missing required 'events' claim")
	}

	if _, exists := claims.Events[backChannelLogoutEvent]; !exists {
		return fmt.Errorf("missing required back-channel logout event in 'events' claim")
	}

	return nil
}

func validateIssuer(claims *oidc.LogoutTokenClaims, feServer *server.FEServer, log zerolog.Logger) error {
	if feServer == nil || feServer.RelayingParty == nil {
		return nil
	}

	expectedIssuer := feServer.RelayingParty.Issuer()
	if claims.Issuer != expectedIssuer {
		log.Warn().
			Str("expected_issuer", expectedIssuer).
			Str("token_issuer", claims.Issuer).
			Msg("Issuer mismatch in logout token")
		return fmt.Errorf("issuer mismatch: expected %s, got %s", expectedIssuer, claims.Issuer)
	}

	return nil
}

func validateJTIUniqueness(claims *oidc.LogoutTokenClaims, feServer *server.FEServer, log zerolog.Logger) error {
	if feServer == nil || claims.JWTID == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), jtiCheckTimeout)
	defer cancel()

	jtiFilter := map[string]interface{}{
		"jti": claims.JWTID,
		"created_at": map[string]interface{}{
			"$gte": time.Now().Add(-jtiReplayCheckWindow),
		},
	}

	count, err := feServer.DB.Collection(logoutTokenJtisCollection).CountDocuments(ctx, jtiFilter)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to check JTI uniqueness")
		return nil
	}

	if count > 0 {
		return fmt.Errorf("logout token replay detected: JTI %s already used", claims.JWTID)
	}

	storeJTIForReplayProtection(claims, feServer, log)
	return nil
}

func storeJTIForReplayProtection(claims *oidc.LogoutTokenClaims, feServer *server.FEServer, log zerolog.Logger) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), sessionCleanupTimeout)
		defer cancel()

		jtiDoc := map[string]interface{}{
			"jti":        claims.JWTID,
			"issuer":     claims.Issuer,
			"subject":    claims.Subject,
			"session_id": claims.SessionID,
			"created_at": time.Now(),
			"expires_at": claims.Expiration.AsTime(),
		}

		_, err := feServer.DB.Collection(logoutTokenJtisCollection).InsertOne(ctx, jtiDoc)
		if err != nil {
			log.Error().Err(err).Msg("Failed to store JTI for replay protection")
		}
	}()
}

func buildUserInfoCallback(resourceServer rs.ResourceServer) rp.CodeExchangeUserinfoCallback[*oidc.IDTokenClaims, *oidc.UserInfo] {
	return func(
		w http.ResponseWriter,
		r *http.Request,
		tokens *oidc.Tokens[*oidc.IDTokenClaims],
		sessionState string,
		provider rp.RelyingParty,
		info *oidc.UserInfo,
	) {
		log := logger.GetLogger(r.Context(), "auth")

		feServer, err := server.ExtractFEServer(r.Context())
		if err != nil {
			log.Error().Err(err).Msg("Failed to extract FEServer from context")
			http.Error(w, internalServerError, http.StatusInternalServerError)
			return
		}

		// resp, err := rs.Introspect[*oidc.IntrospectionResponse](r.Context(), resourceServer, tokens.AccessToken)
		// if err != nil {
		// 	http.Error(w, internalServerError, http.StatusInternalServerError)
		// 	return
		// }

		// // TODO resp.Claims["resource_access"].(map[string][]string)
		userInfo := auth.Convert(info, nil)
		sessionStateBytes, _ := base64.URLEncoding.DecodeString(sessionState)
		sessionState = string(sessionStateBytes)
		log.Debug().
			Dict("tokens", zerolog.Dict().
				Str("issuer", tokens.IDTokenClaims.Issuer).
				Str("subject", tokens.IDTokenClaims.Subject).
				Str("session_id", tokens.IDTokenClaims.SessionID)).
			Str("session_state", sessionState).
			Any("user_info", userInfo).
			Msg("On auth callback")

		session, err := createOrGetSession(feServer, r, log)
		if err != nil {
			http.Error(w, internalServerError, http.StatusInternalServerError)
			return
		}

		populateSessionValues(session, tokens, userInfo, sessionState)

		if err := session.Save(r, w); err != nil {
			log.Error().Err(err).Msg("Failed to save session")
			http.Error(w, internalServerError, http.StatusInternalServerError)
			return
		}

		log.Trace().Msg("Auth session saved")

		useSessionState := &sessionStateStruct{}
		json.Unmarshal([]byte(sessionState), useSessionState)
		redirectTo := validateRedirectURL(useSessionState.RedirectTo, feServer.HTTPOpts.ServeOptions.ContextRoot)

		http.Redirect(w, r, redirectTo, http.StatusFound)
	}
}

func createOrGetSession(feServer *server.FEServer, r *http.Request, log zerolog.Logger) (*sessions.Session, error) {
	session, err := sessions.GetRegistry(r).Get(feServer.SessionStore, feServer.HTTPOpts.SessionOptions.Name)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create session")
		return nil, err
	}
	return session, nil
}

func populateSessionValues(session *sessions.Session, tokens *oidc.Tokens[*oidc.IDTokenClaims], userInfo *auth.UserInfo, session_state string) {
	session.Values[auth.SessionKeyAuthenticated] = true
	session.Values[auth.SessionKeyIssuer] = tokens.IDTokenClaims.Issuer
	session.Values[auth.SessionKeySubject] = tokens.IDTokenClaims.Subject
	session.Values[auth.SessionKeySessionID] = tokens.IDTokenClaims.SessionID
	session.Values[auth.SessionKeyIDToken] = tokens.IDToken
	session.Values[auth.SessionKeyUserInfo] = userInfo
	session.Values[auth.SessionKeyAccessToken] = tokens.AccessToken
	session.Values[auth.SessionKeyRefreshToken] = tokens.RefreshToken
	session.Values[auth.SessionKeyExpiresIn] = tokens.ExpiresIn
	session.Values[auth.SessionKeyJTI] = tokens.IDTokenClaims.JWTID
	session.Values[auth.SessionKeyExpiresAt] = tokens.IDTokenClaims.Expiration.AsTime()
	session.Values[auth.SessionKeySessionState] = session_state
	session.Values[auth.SessionKeyTokenIssuedAt] = time.Now()
}

func extractLogoutToken(r *http.Request, log zerolog.Logger) (string, error) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read request body")
		return "", fmt.Errorf("failed to read request body")
	}
	defer r.Body.Close()

	skip := len(logoutTokenParam)
	if len(bodyBytes) < skip {
		return "", fmt.Errorf("invalid request body")
	}
	if string(bodyBytes[:skip]) != logoutTokenParam {
		return "", fmt.Errorf("invalid request body")
	}
	if len(bodyBytes) == skip {
		return "", fmt.Errorf("empty logout token")
	}

	return string(bodyBytes[skip:]), nil
}

func parseLogoutToken(tokenStr string, log zerolog.Logger) (*oidc.LogoutTokenClaims, error) {
	claims := &oidc.LogoutTokenClaims{}
	_, err := oidc.ParseToken(tokenStr, claims)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse logout_token JWT")
		return nil, err
	}
	return claims, nil
}

func cleanupUserSessions(claims *oidc.LogoutTokenClaims, feServer *server.FEServer, log zerolog.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), sessionCleanupTimeout)
	defer cancel()

	log.Debug().
		Interface("claims", claims).
		Msg("Cleaning up user sessions in back channel logout")

	filter := createFilterForLogout(claims.Issuer, claims.Subject, claims.SessionID)
	res, err := feServer.DB.Collection(httpSessionsCollection).DeleteMany(ctx, filter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete sessions in back channel logout")
		return
	}

	log.Debug().
		Int64("deleted_count", res.DeletedCount).
		Msg("Deleted sessions in back channel logout")
}

func createFilterForLogout(issuer string, subject string, sessionID string) map[string]interface{} {
	filter := make(map[string]interface{})
	if issuer != "" {
		filter["iam_issuer"] = issuer
	}
	if subject != "" {
		filter["iam_subject"] = subject
	}
	if sessionID != "" {
		filter["iam_session_id"] = sessionID
	}
	return filter
}
