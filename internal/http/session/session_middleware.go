package session

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/morphy76/g-fe-server/internal/logger"
)

// BindHTTPSessionToRequests injects the session store into the request context
func BindHTTPSessionToRequests(sessionStore sessions.Store, sessionName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, _ := sessionStore.Get(r, sessionName)
			wrapper := NewWrapper(session)
			defer func() {
				if wrapper.IsDirty() {
					err := session.Save(r, w)
					if err != nil {
						useLogger := logger.GetLogger(r.Context(), "http")
						useLogger.Error().Err(err).Msg("Failed to save session")
					}
				}
			}()

			sessionContext := InjectSession(r.Context(), wrapper)
			useRequest := r.WithContext(sessionContext)

			next.ServeHTTP(w, useRequest)
		})
	}
}
