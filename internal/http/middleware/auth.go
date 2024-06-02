package middleware

import (
	"fmt"
	"net/http"

	app_http "github.com/morphy76/g-fe-server/internal/http"
)

func AuthenticationRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveOptions := app_http.ExtractServeOptions(r.Context())
		session := app_http.ExtractSession(r.Context())
		logger := app_http.ExtractLogger(r.Context(), "auth")

		authURL := fmt.Sprintf(
			"%s://%s:%s/%s/auth/login",
			serveOptions.Protocol,
			serveOptions.Host,
			serveOptions.Port,
			serveOptions.ContextRoot,
		)

		idToken := session.Values["id_token"]
		if idToken == nil {
			logger.Trace().Msg("Redirecting to login")
			http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
}
