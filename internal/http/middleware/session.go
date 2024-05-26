package middleware

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

func InjectSession(router *mux.Router, store sessions.Store) {
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			session, _ := store.Get(r, "http_session")
			defer session.Save(r, w)

			next.ServeHTTP(w, r)
		})
	})
}
