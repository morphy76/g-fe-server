package middleware

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/rs/zerolog/log"

	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/morphy76/g-fe-server/internal/options"
)

func InjectSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		store := r.Context().Value(app_http.CTX_SESSION_STORE_KEY).(sessions.Store)
		serveOptions := r.Context().Value(app_http.CTX_CONTEXT_SERVE_KEY).(*options.ServeOptions)

		session, _ := store.Get(r, serveOptions.SessionName)

		log.Trace().
			Bool("new session", session.IsNew).
			Str("session name", session.Name()).
			Msg("Session injected")

		sessionContext := context.WithValue(r.Context(), app_http.CTX_SESSION_KEY, session)
		useRequest := r.WithContext(sessionContext)

		next.ServeHTTP(w, useRequest)
	})
}
