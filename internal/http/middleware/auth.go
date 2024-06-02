package middleware

import (
	"fmt"
	"net/http"
	"net/url"

	app_http "github.com/morphy76/g-fe-server/internal/http"
)

func AuthenticationRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveOptions := app_http.ExtractServeOptions(r.Context())
		session := app_http.ExtractSession(r.Context())
		logger := app_http.ExtractLogger(r.Context(), "auth")

		authURL := fmt.Sprintf(
			"%s://%s:%s/%s/auth/login?requested_url=%s",
			serveOptions.Protocol,
			serveOptions.Host,
			serveOptions.Port,
			serveOptions.ContextRoot,
			url.QueryEscape(r.URL.String()),
		)

		idToken := session.Values["id_token"]
		if idToken == nil || len(idToken.(string)) == 0 {
			logger.Trace().
				Str("requested_url", r.URL.String()).
				Msg("Redirecting to login")
			w.Header().Set("Cache-Control", "no-cache")
			http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
			return
		}

		next.ServeHTTP(w, r)
	})
}
