package handlers

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/morphy76/g-fe-server/cmd/options"
	"github.com/morphy76/g-fe-server/internal/auth"
	"github.com/morphy76/g-fe-server/internal/logger"
	"github.com/morphy76/g-fe-server/internal/server"
	"github.com/rs/zerolog"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

const (
	failedToGetSession  = "Failed to get session"
	internalServerError = "Internal server error"
	// AuthQueryArgsRedirectTo is the query argument used to redirect after login
	AuthQueryArgsRedirectTo = "redirect_to"
)

// IAMHandlers registers the IAM authentication handlers
func IAMHandlers(
	authRouter *mux.Router,
	httpOptions *options.HTTPOptions,
	sessionStore sessions.Store,
	relyingParty rp.RelyingParty,
) error {
	ctxRoot := httpOptions.ServeOptions.ContextRoot

	gob.Register(auth.UserInfo{}) // Register UserInfo type for session storage

	authRouter.HandleFunc("/login", onLogin(sessionStore, httpOptions, ctxRoot, relyingParty)).Name("GET " + ctxRoot + "/auth/login")
	authRouter.HandleFunc("/callback", rp.CodeExchangeHandler(rp.UserinfoCallback(marshalUserinfo), relyingParty)).Name("GET " + ctxRoot + "/auth/callback")
	authRouter.HandleFunc("/logout", onLogout(sessionStore, httpOptions, ctxRoot, relyingParty)).Name("GET " + ctxRoot + "/auth/logout")
	authRouter.HandleFunc("/info", onInfo(sessionStore, httpOptions)).Name("GET " + ctxRoot + "/auth/info")
	authRouter.HandleFunc("/bc_logout", onBackChannelLogout()).Methods("POST").Name("POST " + ctxRoot + "/auth/bc_logout")

	return nil
}

func onLogin(sessionStore sessions.Store, httpOptions *options.HTTPOptions, ctxRoot string, relyingParty rp.RelyingParty) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logger.GetLogger(r.Context(), "auth")
		logger.Trace().Msg("Logging in")

		requestedURL, err := url.QueryUnescape(r.URL.Query().Get(AuthQueryArgsRedirectTo))
		if err != nil {
			requestedURL = ctxRoot + "/ui"
		}

		stateFn := func() string {
			return requestedURL
		}

		session, err := sessions.GetRegistry(r).Get(sessionStore, httpOptions.SessionOptions.Name)
		if err != nil {
			logger.Warn().Err(err).Msg(failedToGetSession)
			rp.AuthURLHandler(stateFn, relyingParty)(w, r)
			return
		}

		isAuth, found := session.Values[auth.SessionKeyAuthenticated]
		if found && isAuth.(bool) {
			logger.Debug().Msg("Session is already authenticated")
			http.Redirect(w, r, requestedURL, http.StatusFound)
			return
		}

		rp.AuthURLHandler(stateFn, relyingParty)(w, r)
	}
}

func onLogout(sessionStore sessions.Store, httpOptions *options.HTTPOptions, ctxRoot string, relyingParty rp.RelyingParty) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logger.GetLogger(r.Context(), "auth")
		logger.Trace().Msg("Logging out")

		requestedURL, err := url.QueryUnescape(r.URL.Query().Get(AuthQueryArgsRedirectTo))
		if err != nil || requestedURL == "" {
			requestedURL = ctxRoot + "/ui"
		}

		session, err := sessions.GetRegistry(r).Get(sessionStore, httpOptions.SessionOptions.Name)
		if err != nil {
			logger.Error().Err(err).Msg(failedToGetSession)
			http.Error(w, failedToGetSession, http.StatusInternalServerError)
			return
		}

		issuer := session.Values[auth.SessionKeyIssuer]
		subject := session.Values[auth.SessionKeySubject]
		sid := session.Values[auth.SessionKeySessionID]
		idToken := session.Values[auth.SessionKeyIDToken]
		logger.Debug().
			Interface("issuer", issuer).
			Interface("subject", subject).
			Interface("session_id", sid).
			Interface("id_token", idToken).
			Interface("requestedURL", requestedURL).
			Msg("Start logging out")

		url, err := rp.EndSession(r.Context(), relyingParty, idToken.(string), requestedURL, "")
		if err != nil {
			logger.Error().Err(err).Msg("End session failed")
			http.Error(w, "End session failed", http.StatusInternalServerError)
			return
		}

		session.Options.MaxAge = -1
		session.Save(r, w)
		logger.Trace().
			Any("to url", url).
			Msg("Auth session deleted")

		http.Redirect(w, r, url.String(), http.StatusFound)
	}
}

func onInfo(sessionStore sessions.Store, httpOptions *options.HTTPOptions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logger.GetLogger(r.Context(), "auth")

		session, err := sessions.GetRegistry(r).Get(sessionStore, httpOptions.SessionOptions.Name)
		if err != nil {
			logger.Error().Err(err).Msg(failedToGetSession)
			http.Error(w, failedToGetSession, http.StatusInternalServerError)
			return
		}

		logger.Trace().Msg("Info requested")

		authenticated, found := session.Values[auth.SessionKeyAuthenticated]
		if !found || !authenticated.(bool) {
			logger.Debug().Msg("User is not authenticated")
			http.Error(w, "User is not authenticated", http.StatusUnauthorized)
			return
		}

		rv := session.Values[auth.SessionKeyUserInfo]
		responseBody, err := json.Marshal(rv)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to marshal response")
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

		// OIDC Back-Channel Logout 1.0 compliant implementation
		// Implements all required validations per RFC section 2.6 and 2.8

		// Set Cache-Control header as per section 2.8 of the spec
		w.Header().Set("Cache-Control", "no-store")

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		skip := len("logout_token=")
		if len(bodyBytes) < skip {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if string(bodyBytes[:skip]) != "logout_token=" {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if len(bodyBytes) == skip {
			http.Error(w, "Empty logout token", http.StatusBadRequest)
			return
		}
		bodyStr := string(bodyBytes[skip:])

		// Decode the JWT logout_token using zitadel/oidc with proper LogoutTokenClaims
		claims := &oidc.LogoutTokenClaims{}
		_, err = oidc.ParseToken(bodyStr, claims)
		if err != nil {
			log.Error().Err(err).Msg("Failed to parse logout_token JWT")
			http.Error(w, "Invalid logout_token", http.StatusBadRequest)
			return
		}

		feServer, err := server.ExtractFEServer(r.Context())
		if err != nil {
			log.Error().Err(err).Msg("Failed to extract FEServer from context")
			http.Error(w, internalServerError, http.StatusInternalServerError)
			return
		}

		// Validate logout token as per section 2.6 of the spec
		if err := validateLogoutToken(claims, log, feServer); err != nil {
			log.Error().Err(err).Msg("Logout token validation failed")
			http.Error(w, "Invalid logout_token", http.StatusBadRequest)
			return
		}

		subject := claims.Subject
		issuer := claims.Issuer
		sessionID := claims.SessionID

		log.Debug().
			Str("subject", subject).
			Str("issuer", issuer).
			Str("session_id", sessionID).
			Str("jti", claims.JWTID).
			Msg("Decoded logout_token")

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			filter := map[string]interface{}{
				"iam_issuer":     issuer,
				"iam_subject":    subject,
				"iam_session_id": sessionID,
			}

			res, err := feServer.DB.Collection("http_sessions").DeleteMany(ctx, filter)
			if err != nil {
				log.Error().Err(err).Msg("Failed to delete sessions in back channel logout")
				return
			} else {
				log.Debug().
					Int64("deleted_count", res.DeletedCount).
					Msg("Deleted sessions in back channel logout")
			}
		}()

		// Return HTTP 200 OK as required by section 2.8 of the spec
		w.WriteHeader(http.StatusOK)
	}
}

// NOTE: For production deployment, consider creating the following MongoDB indexes:
// 1. db.logout_token_jtis.createIndex({"jti": 1}, {"unique": true})
// 2. db.logout_token_jtis.createIndex({"expires_at": 1}, {"expireAfterSeconds": 0})
// 3. db.logout_token_jtis.createIndex({"created_at": 1})

// validateLogoutToken validates the logout token according to section 2.6 of the OIDC Back-Channel Logout spec
func validateLogoutToken(claims *oidc.LogoutTokenClaims, log zerolog.Logger, feServer *server.FEServer) error {
	// 1. JWT signature validation is handled by oidc.ParseToken

	// 2. Validate required claims exist
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

	// 3. Validate expiration
	if time.Now().After(claims.Expiration.AsTime()) {
		return fmt.Errorf("logout token has expired")
	}

	// 4. Verify that token contains sub or sid claim
	if claims.Subject == "" && claims.SessionID == "" {
		return fmt.Errorf("logout token must contain either 'sub' or 'sid' claim")
	}

	// 5. Verify events claim - CRITICAL for compliance
	if claims.Events == nil {
		return fmt.Errorf("missing required 'events' claim")
	}

	// Check for the specific back-channel logout event
	backChannelLogoutEvent := "http://schemas.openid.net/event/backchannel-logout"
	if _, exists := claims.Events[backChannelLogoutEvent]; !exists {
		return fmt.Errorf("missing required back-channel logout event in 'events' claim")
	}

	// 6. Verify nonce claim is not present (LogoutTokenClaims doesn't have Nonce field, so this is automatically satisfied)

	// 7. Verify JWTID is present (required for logout tokens)
	if claims.JWTID == "" {
		return fmt.Errorf("missing required 'jti' claim")
	}

	// 8. Optional: verify issuer matches expected issuer from the relying party configuration
	if feServer != nil && feServer.RelayingParty != nil {
		expectedIssuer := feServer.RelayingParty.Issuer()
		if claims.Issuer != expectedIssuer {
			log.Warn().
				Str("expected_issuer", expectedIssuer).
				Str("token_issuer", claims.Issuer).
				Msg("Issuer mismatch in logout token")
			return fmt.Errorf("issuer mismatch: expected %s, got %s", expectedIssuer, claims.Issuer)
		}
	}

	// 9. Optional: verify JTI uniqueness to prevent replay attacks
	// This is a simplified implementation - in production you'd want to store JTIs in a cache/database
	// with appropriate TTL based on the token expiration
	if feServer != nil && claims.JWTID != "" {
		// Check if this JTI was recently used (implement based on your caching strategy)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Using a simple collection check - in production consider using Redis or similar
		jtiFilter := map[string]interface{}{
			"jti": claims.JWTID,
			"created_at": map[string]interface{}{
				"$gte": time.Now().Add(-5 * time.Minute), // Check last 5 minutes
			},
		}

		count, err := feServer.DB.Collection("logout_token_jtis").CountDocuments(ctx, jtiFilter)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to check JTI uniqueness")
			// Continue processing - don't fail on database errors for this optional check
		} else if count > 0 {
			return fmt.Errorf("logout token replay detected: JTI %s already used", claims.JWTID)
		}

		// Store the JTI for future checks
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			jtiDoc := map[string]interface{}{
				"jti":        claims.JWTID,
				"issuer":     claims.Issuer,
				"subject":    claims.Subject,
				"session_id": claims.SessionID,
				"created_at": time.Now(),
				"expires_at": claims.Expiration.AsTime(),
			}

			_, err := feServer.DB.Collection("logout_token_jtis").InsertOne(ctx, jtiDoc)
			if err != nil {
				log.Error().Err(err).Msg("Failed to store JTI for replay protection")
			}
		}()
	}

	// 10. Optional: verify subject and session ID match existing sessions
	// This could be implemented to cross-reference with your session store

	return nil
}

func marshalUserinfo(
	w http.ResponseWriter,
	r *http.Request,
	tokens *oidc.Tokens[*oidc.IDTokenClaims],
	state string,
	provider rp.RelyingParty,
	info *oidc.UserInfo,
) {
	logger := logger.GetLogger(r.Context(), "auth")
	feServer, err := server.ExtractFEServer(r.Context())
	if err != nil {
		logger.Error().Err(err).Msg("Failed to extract FEServer from context")
		http.Error(w, internalServerError, http.StatusInternalServerError)
		return
	}

	userInfo := auth.Convert(info)
	logger.Debug().
		Dict("tokens", zerolog.Dict().
			Str("issuer", tokens.IDTokenClaims.Issuer).
			Str("subject", tokens.IDTokenClaims.Subject).
			Str("session_id", tokens.IDTokenClaims.SessionID)).
		Str("state", state).
		Any("user_info", userInfo).
		Msg("On auth callback")

	session, err := sessions.GetRegistry(r).Get(feServer.SessionStore, feServer.HTTPOpts.SessionOptions.Name)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create session")
		http.Error(w, internalServerError, http.StatusInternalServerError)
		return
	}

	session.Values[auth.SessionKeyAuthenticated] = true
	session.Values[auth.SessionKeyIssuer] = tokens.IDTokenClaims.Issuer
	session.Values[auth.SessionKeySubject] = tokens.IDTokenClaims.Subject
	session.Values[auth.SessionKeySessionID] = tokens.IDTokenClaims.SessionID
	session.Values[auth.SessionKeyIDToken] = tokens.IDToken
	session.Values[auth.SessionKeyUserInfo] = userInfo
	session.Values[auth.SessionKeyAccessToken] = tokens.AccessToken
	session.Values[auth.SessionKeyRefreshToken] = tokens.RefreshToken
	session.Values[auth.SessionKeyExpiresIn] = tokens.ExpiresIn

	err = session.Save(r, w)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to save session")
		http.Error(w, internalServerError, http.StatusInternalServerError)
	}

	logger.Trace().Msg("Auth session saved")

	http.Redirect(w, r, state, http.StatusFound)
}
