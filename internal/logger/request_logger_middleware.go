package logger

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
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

		hook := zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
			// activeSpan := trace.SpanFromContext(r.Context())
			// if activeSpan.SpanContext().TraceID().IsValid() {
			// 	e.Dict("correlation", zerolog.Dict().
			// 		Str("span_id", activeSpan.SpanContext().SpanID().String()).
			// 		Str("trace_id", activeSpan.SpanContext().TraceID().String()),
			// 	)
			// }
		})
		requestLogger := GetLogger(r.Context(), "http").Hook(hook)

		before := time.Now()
		next.ServeHTTP(recorder, r)
		requestDuration := time.Since(before)

		if requestLogger.Trace().Enabled() {
			requestLogger.Trace().Dict("headers", dumpHeaders(r.Header)).Msg("Request Header")
		}

		requestLogger.Debug().Dict("request", zerolog.Dict().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Dur("duration_ns", time.Duration(requestDuration)).
			Int("code", recorder.Status),
		).Msg("HTTP Request")

		if requestLogger.Trace().Enabled() {
			requestLogger.Trace().Dict("headers", dumpHeaders(recorder.ResponseWriter.Header())).Msg("Response Header")
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
