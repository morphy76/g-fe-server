package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type statusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

type CTX_LOGGER string

const CTX_LOGGER_KEY CTX_LOGGER = "logger"

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{
			ResponseWriter: w,
			Status:         200,
		}

		ownership := r.Context().Value(CTX_OWNERSHIP_KEY).(Ownership)
		useLogger := log.Logger.With().
			Dict("ownership", zerolog.Dict().
				Str("tenant", ownership.Tenant).
				Str("subscription", ownership.Subscription),
			).Logger()
		newContext := context.WithValue(r.Context(), CTX_LOGGER_KEY, useLogger)
		useRequestLogger := r.WithContext(newContext)

		start := time.Now()
		next.ServeHTTP(recorder, useRequestLogger)
		elapsed := time.Since(start)

		useLogger.Debug().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("code", recorder.Status).
			Int64("duration_ms", elapsed.Microseconds()).
			Msg("HTTP Request")
	})
}
