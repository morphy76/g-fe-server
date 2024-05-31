package auth

import (
	"net/http"

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
		session.Values["email"] = info.Email
		session.Values["family_name"] = info.FamilyName
		session.Values["given_name"] = info.GivenName
		session.Values["name"] = info.Name
		session.Values["preferred_username"] = info.PreferredUsername

		session.Save(r, w)

		http.Redirect(w, r, ctxRoot+"/ui", http.StatusFound)
	}

	authRouter.HandleFunc("/login", rp.AuthURLHandler(todo, relyingParty)).Methods("GET").Name(ctxRoot + "/auth/login")
	authRouter.HandleFunc("/callback", rp.CodeExchangeHandler(rp.UserinfoCallback(marshalUserinfo), relyingParty)).Name(ctxRoot + "/auth/callback")
}
