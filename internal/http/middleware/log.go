package middleware

import (
	"net/http"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"

	app_http "github.com/morphy76/g-fe-server/internal/http"
)

type statusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{
			ResponseWriter: w,
			Status:         200,
		}

		activeSpan := trace.SpanFromContext(r.Context())

		ownership := app_http.ExtractOwnership(r.Context())
		useLogger := log.Logger.With().
			Dict("correlation", zerolog.Dict().
				Str("span_id", activeSpan.SpanContext().SpanID().String()).
				Str("trace_id", activeSpan.SpanContext().TraceID().String()),
			).
			Dict("ownership", zerolog.Dict().
				Str("tenant", ownership.Tenant).
				Str("subscription", ownership.Subscription),
			).Logger()
		newContext := app_http.InjectLogger(r.Context(), useLogger)
		useRequestLogger := r.WithContext(newContext)

		next.ServeHTTP(recorder, useRequestLogger)

		useLogger.Debug().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("code", recorder.Status).
			Msg("HTTP Request")
	})
}
