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

		backTo := r.URL.String()
		backTo = url.QueryEscape(backTo)

		authURL := fmt.Sprintf(
			"%s://%s:%s/%s/auth/login?backTo=%s",
			serveOptions.Protocol,
			serveOptions.Host,
			serveOptions.Port,
			serveOptions.ContextRoot,
			backTo,
		)

		idToken := session.Values["id_token"]
		if idToken == nil || len(idToken.(string)) == 0 {
			http.Redirect(w, r, authURL, http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}
