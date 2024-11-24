package middleware

import (
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/morphy76/g-fe-server/internal/logger"
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

		// activeSpan := trace.SpanFromContext(r.Context())
		// ownership := app_http.ExtractOwnership(r.Context())

		// useLoggerBuilder := log.Logger.With().
		// 	Dict("correlation", zerolog.Dict().
		// 		Str("span_id", activeSpan.SpanContext().SpanID().String()).
		// 		Str("trace_id", activeSpan.SpanContext().TraceID().String()),
		// 	)
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
		// useLogger := useLoggerBuilder.Logger()

		// newContext := app_http.InjectLogger(r.Context(), useLogger)
		// useRequestLogger := r.WithContext(newContext)

		requestLogger := logger.GetLogger(r.Context(), "http")
		next.ServeHTTP(recorder, r)

		if requestLogger.Trace().Enabled() {
			for k, v := range r.Header {
				log.Trace().Strs(k, v).Msg("Header")
			}
		}

		requestLogger.Debug().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("code", recorder.Status).
			Msg("HTTP Request")

		if requestLogger.Trace().Enabled() {
			for k, v := range recorder.ResponseWriter.Header() {
				log.Trace().Strs(k, v).Msg("Response Header")
			}
		}
	})
}
