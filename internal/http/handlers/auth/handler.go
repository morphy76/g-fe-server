package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"

	app_http "github.com/morphy76/g-fe-server/internal/http"
)

func IAMHandlers(authRouter *mux.Router, ctxRoot string, relyingParty rp.RelyingParty) {

	stateFn := func() string {
		return uuid.New().String()
	}

	marshalUserinfo := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty, info *oidc.UserInfo) {

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

		backTo := ctxRoot + "/ui"
		// sessionBackTo := tokens.IDTokenClaims.Claims["session_state"]
		// if sessionBackTo != nil {
		// 	backTo = sessionBackTo.(string)
		// }

		http.Redirect(w, r, backTo, http.StatusFound)
	}

	authRouter.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {

		relyingParty := app_http.ExtractRelyingParty(r.Context())
		logger := app_http.ExtractLogger(r.Context(), "auth")

		authURL := rp.AuthURL(stateFn(), relyingParty)
		logger.Trace().
			Str("auth_url", authURL).
			Msg("Redirecting to auth")

		http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)

	}).Methods("GET").Name("GET " + ctxRoot + "/auth/login")

	authRouter.HandleFunc("/callback", rp.CodeExchangeHandler(rp.UserinfoCallback(marshalUserinfo), relyingParty)).Name("GET " + ctxRoot + "/auth/callback")

	authRouter.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {

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

			http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
			return
		}

		sessionState := session.Values["session_state"].(string)

		session.Options.MaxAge = -1
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

		http.Redirect(w, r, url.String(), http.StatusTemporaryRedirect)

	}).Name("GET " + ctxRoot + "/auth/logout")

	authRouter.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {

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
	}).Name("GET " + ctxRoot + "/auth/info")
}
