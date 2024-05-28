package middleware

import (
	"context"
	"net/http"
	"time"

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

		ownership := r.Context().Value(app_http.CTX_OWNERSHIP_KEY).(Ownership)
		useLogger := log.Logger.With().
			Dict("correlation", zerolog.Dict().
				Str("span_id", activeSpan.SpanContext().SpanID().String()).
				Str("trace_id", activeSpan.SpanContext().TraceID().String()),
			).
			Dict("ownership", zerolog.Dict().
				Str("tenant", ownership.Tenant).
				Str("subscription", ownership.Subscription),
			).Logger()
		newContext := context.WithValue(r.Context(), app_http.CTX_LOGGER_KEY, useLogger)
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

func ExtractLoggerFromContext(ctx context.Context, forPackage string) zerolog.Logger {
	return (ctx.Value(app_http.CTX_LOGGER_KEY).(zerolog.Logger)).With().Str("package", forPackage).Logger()
}

func ExtractLoggerFromRequest(r *http.Request, forPackage string) zerolog.Logger {
	return ExtractLoggerFromContext(r.Context(), forPackage)
}
