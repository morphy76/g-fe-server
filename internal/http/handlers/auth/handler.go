package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"

	app_http "github.com/morphy76/g-fe-server/internal/http"
)

func IAMHandlers(authRouter *mux.Router, ctxRoot string, relyingParty rp.RelyingParty) {
	authRouter.HandleFunc("/login", onLogin(ctxRoot, relyingParty)).Methods("GET").Name("GET " + ctxRoot + "/auth/login")
	authRouter.HandleFunc("/callback", rp.CodeExchangeHandler(rp.UserinfoCallback(marshalUserinfo), relyingParty)).Name("GET " + ctxRoot + "/auth/callback")
	authRouter.HandleFunc("/logout", onLogout()).Name("GET " + ctxRoot + "/auth/logout")
	authRouter.HandleFunc("/info", onInfo(ctxRoot)).Name("GET " + ctxRoot + "/auth/info")
}

func onLogin(ctxRoot string, relyingParty rp.RelyingParty) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		requested_url, err := url.QueryUnescape(r.URL.Query().Get("requested_url"))
		if err != nil {
			requested_url = ctxRoot + "/ui"
		}

		stateFn := func() string {
			return requested_url
		}

		rp.AuthURLHandler(stateFn, relyingParty)(w, r)
	}
}

func onLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		logger := app_http.ExtractLogger(r.Context(), "auth")
		logger.Trace().Msg("Logging out")

		session := app_http.ExtractSession(r.Context())
		serveOptions := app_http.ExtractServeOptions(r.Context())
		useRelyingParty := app_http.ExtractRelyingParty(r.Context())

		logger.Trace().Msg("Start logging out")

		backTo := fmt.Sprintf(
			"%s://%s:%s/%s/ui/",
			serveOptions.Protocol,
			serveOptions.Host,
			serveOptions.Port,
			serveOptions.ContextRoot,
		)

		idToken := session.Values["id_token"]
		if idToken == nil {
			authURL := fmt.Sprintf(
				"%s://%s:%s/%s/auth/login",
				serveOptions.Protocol,
				serveOptions.Host,
				serveOptions.Port,
				serveOptions.ContextRoot,
			)

			w.Header().Set("Cache-Control", "no-cache")
			http.Redirect(w, r, authURL, http.StatusFound)
			return
		}

		sessionState := session.Values["session_state"].(string)

		session.Options.MaxAge = -1
		delete(session.Values, "id_token")
		session.Save(r, w)
		url, err := rp.EndSession(context.Background(), useRelyingParty, idToken.(string), backTo, sessionState)
		if err != nil {
			logger.Error().Err(err).Msg("End session failed")
			http.Error(w, "End session failed", http.StatusInternalServerError)
			return
		}
		logger.Trace().
			Any("to url", url).
			Msg("Auth session deleted")

		w.Header().Set("Cache-Control", "no-cache")
		http.Redirect(w, r, url.String(), http.StatusFound)
	}
}

func onInfo(ctxRoot string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		session := app_http.ExtractSession(r.Context())
		logger := app_http.ExtractLogger(r.Context(), "auth")

		logger.Trace().Msg("Info requested")

		idToken := session.Values["id_token"]
		if idToken == nil {
			http.Error(w, "Auth session not found", http.StatusUnauthorized)
			return
		}

		rv := &map[string]string{
			"email":              session.Values["email"].(string),
			"family_name":        session.Values["family_name"].(string),
			"given_name":         session.Values["given_name"].(string),
			"name":               session.Values["name"].(string),
			"preferred_username": session.Values["preferred_username"].(string),
			"logout_url":         ctxRoot + "/auth/logout",
		}
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

func marshalUserinfo(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty, info *oidc.UserInfo) {

	session := app_http.ExtractSession(r.Context())
	logger := app_http.ExtractLogger(r.Context(), "auth")

	session.Values["access_token"] = tokens.AccessToken
	session.Values["refresh_token"] = tokens.RefreshToken
	session.Values["id_token"] = tokens.IDToken
	session.Values["session_state"] = tokens.IDTokenClaims.Claims["session_state"]
	session.Values["email"] = info.Email
	session.Values["family_name"] = info.FamilyName
	session.Values["given_name"] = info.GivenName
	session.Values["name"] = info.Name
	session.Values["preferred_username"] = info.PreferredUsername

	session.Save(r, w)
	logger.Trace().Msg("Auth session saved")

	http.Redirect(w, r, state, http.StatusFound)
}
