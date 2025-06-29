package handlers

import (
	"context"
	"encoding/gob"
	"encoding/json"
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

		feServer, err := server.ExtractFEServer(r.Context())
		if err != nil {
			log.Error().Err(err).Msg("Failed to extract FEServer from context")
			http.Error(w, internalServerError, http.StatusInternalServerError)
			return
		}

		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			res, err := feServer.DB.Collection("http_sessions").DeleteMany(ctx, map[string]interface{}{})
			if err != nil {
				log.Error().Err(err).Msg("Failed to delete sessions in back channel logout")
				return
			} else {
				log.Debug().
					Int64("deleted_count", res.DeletedCount).
					Msg("Deleted sessions in back channel logout")
			}
		}()

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
