package auth

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

func IAMHandlers(authRouter *mux.Router, ctxRoot string, relyingParty rp.RelyingParty) {

	todo := func() string {
		return uuid.New().String()
	}

	marshalUserinfo := func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty, info *oidc.UserInfo) {
		// fmt.Println("access token", tokens.AccessToken)
		// fmt.Println("refresh token", tokens.RefreshToken)
		// fmt.Println("id token", tokens.IDToken)

		data, err := json.Marshal(info)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(data)
	}

	authRouter.HandleFunc("/login", rp.AuthURLHandler(todo, relyingParty)).Methods("GET").Name(ctxRoot + "/auth/login")
	authRouter.HandleFunc("/callback", rp.CodeExchangeHandler(rp.UserinfoCallback(marshalUserinfo), relyingParty)).Name(ctxRoot + "/auth/callback")
}
