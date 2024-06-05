package middleware

import (
	"net/http"

	app_http "github.com/morphy76/g-fe-server/internal/http"
	"github.com/rs/zerolog/log"
)

func InjectSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		store := app_http.ExtractSessionStore(r.Context())
		serveOptions := app_http.ExtractServeOptions(r.Context())

		session, _ := store.Get(r, serveOptions.SessionName)

		log.Trace().
			Bool("new session", session.IsNew).
			Str("session name", session.Name()).
			Msg("Session injected")

		sessionContext := app_http.InjectSession(r.Context(), session)
		useRequest := r.WithContext(sessionContext)

		next.ServeHTTP(w, useRequest)
	})
}
