package middleware

import (
	"net/http"

	app_http "github.com/morphy76/g-fe-server/internal/http"
)

func DumpHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := app_http.ExtractLogger(r.Context(), "troubleshoot")
		for k, v := range r.Header {
			log.Trace().Strs(k, v).Msg("Header")
		}
		next.ServeHTTP(w, r)
	})
}
