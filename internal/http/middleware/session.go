package middleware

import (
	"net/http"

	app_http "github.com/morphy76/g-fe-server/internal/http"
)

func InjectSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		store := app_http.ExtractSessionStore(r.Context())
		serveOptions := app_http.ExtractServeOptions(r.Context())

		session, _ := store.Get(r, serveOptions.SessionName)

		sessionContext := app_http.InjectSession(r.Context(), session)
		useRequest := r.WithContext(sessionContext)

		next.ServeHTTP(w, useRequest)
	})
}
