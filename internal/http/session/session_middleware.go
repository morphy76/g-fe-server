package session

import (
	"net/http"

	"github.com/gorilla/sessions"
)

// BindHTTPSessionToRequests injects the session store into the request context
func BindHTTPSessionToRequests(sessionStore sessions.Store, sessionName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			session, _ := sessionStore.Get(r, sessionName)
			wrapper := NewSessionWrapper(session)

			sessionContext := InjectSession(r.Context(), wrapper)
			useRequest := r.WithContext(sessionContext)

			next.ServeHTTP(w, useRequest)

			if wrapper.IsDirty() {
				session.Save(r, w)
			}
		})
	}
}
