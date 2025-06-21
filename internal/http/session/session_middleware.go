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
			useLogger := logger.GetLogger(r.Context(), "http")

			session, err := sessionStore.Get(r, sessionName)
			if err != nil {
				useLogger.Error().Err(err).Str("session name", sessionName).Msg("Failed to get session")
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			wrapper := NewWrapper(session)
			defer func() {
				if wrapper.IsDirty() {
					err := session.Save(r, w)
					if err != nil {
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
