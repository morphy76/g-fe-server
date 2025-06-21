package handlers

import (
	"encoding/gob"
	"encoding/json"
	"net/http"
	"net/url"

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
	authRouter.HandleFunc("/logout", onLogout(sessionStore, httpOptions, relyingParty)).Name("GET " + ctxRoot + "/auth/logout")
	authRouter.HandleFunc("/info", onInfo(sessionStore, httpOptions, ctxRoot)).Name("GET " + ctxRoot + "/auth/info")
	// authRouter.HandleFunc("/bc_logout", onBackChannelLogout()).Methods("POST").Name("POST " + ctxRoot + "/auth/bc_logout")

	return nil
}

func onLogin(sessionStore sessions.Store, httpOptions *options.HTTPOptions, ctxRoot string, relyingParty rp.RelyingParty) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logger.GetLogger(r.Context(), "auth")
		logger.Trace().Msg("Logging in")

		requestedURL, err := url.QueryUnescape(r.URL.Query().Get("redirect_to"))
		if err != nil {
			requestedURL = ctxRoot + "/ui"
		}

		stateFn := func() string {
			return requestedURL
		}

		session, err := sessions.GetRegistry(r).Get(sessionStore, httpOptions.SessionOptions.Name)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get session")
			rp.AuthURLHandler(stateFn, relyingParty)(w, r)
			return
		}

		isAuth, found := session.Values["authenticated"]
		if found && isAuth.(bool) {
			logger.Debug().Msg("Session is already authenticated")
			http.Redirect(w, r, requestedURL, http.StatusFound)
			return
		}

		rp.AuthURLHandler(stateFn, relyingParty)(w, r)
	}
}

func onLogout(sessionStore sessions.Store, httpOptions *options.HTTPOptions, relyingParty rp.RelyingParty) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logger.GetLogger(r.Context(), "auth")
		logger.Trace().Msg("Logging out")

		session, err := sessions.GetRegistry(r).Get(sessionStore, httpOptions.SessionOptions.Name)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get session")
			// TODO
			return
		}

		issuer := session.Values["issuer"]
		subject := session.Values["subject"]
		sid := session.Values["session_id"]
		idToken := session.Values["id_token"]
		logger.Trace().
			Interface("issuer", issuer).
			Interface("subject", subject).
			Interface("session_id", sid).
			Interface("id_token", idToken).
			Msg("Start logging out")

		session.Options.MaxAge = -1

		url, err := rp.EndSession(r.Context(), relyingParty, idToken.(string), "", "")
		if err != nil {
			logger.Error().Err(err).Msg("End session failed")
			// TODO
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
			logger.Error().Err(err).Msg("Failed to get session")
			http.Error(w, "Failed to get session", http.StatusInternalServerError)
			return
		}

		logger.Trace().Msg("Info requested")

		authenticated, found := session.Values["authenticated"]
		if !found || !authenticated.(bool) {
			logger.Debug().Msg("User is not authenticated")
			http.Error(w, "User is not authenticated", http.StatusUnauthorized)
			return
		}

		rv := session.Values["user_info"]
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

// func onBackChannelLogout() http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		log := logger.GetLogger(r.Context(), "auth")

// 		log.Debug().Msg("Back channel logout")

// 		var body map[string]interface{}
// 		err := json.NewDecoder(r.Body).Decode(&body)
// 		if err != nil {
// 			http.Error(w, "Failed to decode request body", http.StatusBadRequest)
// 			return
// 		}
// 		defer r.Body.Close()

// 		log.Trace().Interface("body", body).Msg("Back channel logout")

// 		w.WriteHeader(http.StatusOK)
// 		w.Write([]byte("OK"))
// 	}
// }

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
		onLogout(feServer.SessionStore, feServer.HTTPOpts, provider)(w, r)
	}

	session.Values["authenticated"] = true
	session.Values["issuer"] = tokens.IDTokenClaims.Issuer
	session.Values["subject"] = tokens.IDTokenClaims.Subject
	session.Values["session_id"] = tokens.IDTokenClaims.SessionID
	session.Values["id_token"] = tokens.IDToken
	session.Values["user_info"] = userInfo
	session.Values["access_token"] = tokens.AccessToken
	session.Values["refresh_token"] = tokens.RefreshToken
	session.Values["expires_in"] = tokens.ExpiresIn

	err = session.Save(r, w)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to save session")
		onLogout(feServer.SessionStore, feServer.HTTPOpts, provider)(w, r)
	}

	logger.Trace().Msg("Auth session saved")

	http.Redirect(w, r, state, http.StatusFound)
}
