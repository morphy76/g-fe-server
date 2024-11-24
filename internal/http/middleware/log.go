package middleware

import (
	"net/http"

	"github.com/morphy76/g-fe-server/internal/logger"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type statusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

// RequestLogger logs the incoming HTTP request
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{
			ResponseWriter: w,
			Status:         200,
		}

		activeSpan := trace.SpanFromContext(r.Context())
		// ownership := app_http.ExtractOwnership(r.Context())

		// if ownership.Tenant != "" {
		// 	useLoggerBuilder = useLoggerBuilder.
		// 		Dict("ownership", zerolog.Dict().
		// 			Str("tenant", ownership.Tenant).
		// 			Str("subscription", ownership.Subscription),
		// 		)
		// } else {
		// 	useLoggerBuilder = useLoggerBuilder.
		// 		Dict("ownership", zerolog.Dict().
		// 			Bool("anon", true),
		// 		)
		// }

		requestLogger := logger.GetLogger(r.Context(), "http").With().
			Dict("correlation", zerolog.Dict().
				Str("span_id", activeSpan.SpanContext().SpanID().String()).
				Str("trace_id", activeSpan.SpanContext().TraceID().String()),
			).
			Logger()
		next.ServeHTTP(recorder, r)

		if requestLogger.Trace().Enabled() {
			requestLogger.Trace().Dict("headers", dumpHeaders(r.Header)).Msg("Request Header")
		}

		requestLogger.Debug().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("code", recorder.Status).
			Msg("HTTP Request")

		if requestLogger.Trace().Enabled() {
			requestLogger.Trace().Dict("headers", dumpHeaders(recorder.ResponseWriter.Header())).Msg("Request Header")
		}
	})
}

func dumpHeaders(headers http.Header) *zerolog.Event {
	events := zerolog.Dict()
	for k, v := range headers {
		events.Strs(k, v)
	}
	return events

}
