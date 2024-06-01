package auth

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"

	app_http "github.com/morphy76/g-fe-server/internal/http"
)

func IAMHandlers(authRouter *mux.Router, ctxRoot string, relyingParty rp.RelyingParty) {

	todo := func() string {
		return uuid.New().String()
	}

	marshalUserinfo := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty, info *oidc.UserInfo) {

		session := app_http.ExtractSession(r.Context())

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

		http.Redirect(w, r, ctxRoot+"/ui", http.StatusFound)
	}
	// relyingParty.OAuthConfig().RedirectURL = "http://localhost:8080/fe/ui/credits"
	authRouter.HandleFunc("/login", rp.AuthURLHandler(todo, relyingParty)).Methods("GET").Name("GET " + ctxRoot + "/auth/login")
	authRouter.HandleFunc("/callback", rp.CodeExchangeHandler(rp.UserinfoCallback(marshalUserinfo), relyingParty)).Name("GET " + ctxRoot + "/auth/callback")
	authRouter.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {

		session := app_http.ExtractSession(r.Context())
		logger := app_http.ExtractLogger(r.Context(), "auth")

		logger.Trace().
			Msg("Start Logout")

		serveOptions := app_http.ExtractServeOptions(r.Context())
		useRelyingParty := app_http.ExtractRelyingParty(r.Context())

		var err error
		backTo := r.URL.Query().Get("backTo")
		if backTo == "" {
			backTo = fmt.Sprintf(
				"%s://%s:%s/%s/ui",
				serveOptions.Protocol,
				serveOptions.Host,
				serveOptions.Port,
				serveOptions.ContextRoot,
			)
		} else {
			backTo, err = url.QueryUnescape(backTo)
			if err != nil {
				logger.Warn().
					Err(err).
					Msg("Failed to unescape backTo")
				backTo = fmt.Sprintf(
					"%s://%s:%s/%s/ui",
					serveOptions.Protocol,
					serveOptions.Host,
					serveOptions.Port,
					serveOptions.ContextRoot,
				)
			}
		}

		logger.Trace().
			Str("client_id", useRelyingParty.OAuthConfig().ClientID).
			Str("redirect_uri", backTo).
			Msg("RP info")

		idToken := session.Values["id_token"]
		if idToken == nil {
			logger.Trace().
				Msg("No ID Token found")

			http.Redirect(w, r, backTo, http.StatusFound)
			return
		}

		sessionState := session.Values["session_state"].(string)

		session.Options.MaxAge = -1
		session.Save(r, w)
		logger.Trace().
			Msg("Session deleted")

		url, err := rp.EndSession(r.Context(), useRelyingParty, idToken.(string), backTo, sessionState)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logger.Trace().
			Str("url", url.String()).
			Msg("Redirect to end session")
		http.Redirect(w, r, url.String(), http.StatusFound)

	}).Name("GET " + ctxRoot + "/auth/logout")
}
