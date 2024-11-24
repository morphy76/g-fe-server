package middleware

import (
	"net/http"

	"github.com/morphy76/g-fe-server/internal/logger"
)

// DumpHeaders logs all headers in the request
func DumpHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.GetLogger(r.Context(), "troubleshoot")
		if log.Trace().Enabled() {
			for k, v := range r.Header {
				log.Trace().Strs(k, v).Msg("Header")
			}
		}
		next.ServeHTTP(w, r)
	})
}
