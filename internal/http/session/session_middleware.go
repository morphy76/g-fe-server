package session

import (
	"net/http"

	"github.com/gorilla/sessions"
)

// BindHTTPSessionToRequests injects the session store into the request context
func BindHTTPSessionToRequests(sessionStore sessions.Store, sessionName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if sessionStore == nil {
				next.ServeHTTP(w, r)
				return
			}

			session, _ := sessionStore.Get(r, sessionName)

			sessionContext := InjectSession(r.Context(), session)
			useRequest := r.WithContext(sessionContext)

			next.ServeHTTP(w, useRequest)
		})
	}
}
