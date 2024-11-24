package middleware

import (
	"net/http"

	"github.com/gorilla/sessions"
	app_http "github.com/morphy76/g-fe-server/internal/http"
)

// InjectSession injects the session store into the request context
func InjectSession(sessionStore sessions.Store, sessionName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			session, _ := sessionStore.Get(r, sessionName)

			sessionContext := app_http.InjectSession(r.Context(), session)
			useRequest := r.WithContext(sessionContext)

			next.ServeHTTP(w, useRequest)
		})
	}
}
