package handlers

import (
	"encoding/gob"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
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
	failedToGetSession = "Failed to get session"
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
	authRouter.HandleFunc("/info", onInfo(sessionStore, httpOptions, ctxRoot)).Name("GET " + ctxRoot + "/auth/info")
	authRouter.HandleFunc("/bc_logout", onBackChannelLogout(sessionStore)).Methods("POST").Name("POST " + ctxRoot + "/auth/bc_logout")

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
		if err != nil {
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
		logger.Trace().
			Interface("issuer", issuer).
			Interface("subject", subject).
			Interface("session_id", sid).
			Interface("id_token", idToken).
			Msg("Start logging out")

		session.Options.MaxAge = -1

		url, err := rp.EndSession(r.Context(), relyingParty, idToken.(string), requestedURL, "")
		if err != nil {
			logger.Error().Err(err).Msg("End session failed")
			http.Error(w, "End session failed", http.StatusInternalServerError)
			return
		}
		logger.Trace().
			Any("to url", url).
			Msg("Auth session deleted")

		http.Redirect(w, r, url.String(), http.StatusFound)
	}
}

func onInfo(sessionStore sessions.Store, httpOptions *options.HTTPOptions, ctxRoot string) http.HandlerFunc {
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

func onBackChannelLogout(sessionStore sessions.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logger.GetLogger(r.Context(), "auth")

		log.Debug().Msg("Back channel logout")

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
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
		defer r.Body.Close()

		token, err := jwt.ParseSigned(bodyStr, []jose.SignatureAlgorithm{
			jose.RS256, // TODO: read from realm configuration
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to parse JWT")
			http.Error(w, "Invalid logout token", http.StatusBadRequest)
			return
		}
		claims := map[string]interface{}{}
		if err := token.UnsafeClaimsWithoutVerification(&claims); err != nil {
			log.Error().Err(err).Msg("Failed to decode JWT claims")
			http.Error(w, "Failed to decode logout token", http.StatusBadRequest)
			return
		}
		log.Info().Interface("claims", claims).Msg("Decoded logout token")
		// TODO delete from sessionStore using iss, sub and/or sid

		w.WriteHeader(http.StatusNoContent)
	}
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}

	logger.Trace().Msg("Auth session saved")

	http.Redirect(w, r, state, http.StatusFound)
}
