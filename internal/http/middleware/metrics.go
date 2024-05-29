package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	app_serve "github.com/morphy76/g-fe-server/internal/serve"
)

func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{
			ResponseWriter: w,
			Status:         200,
		}

		start := time.Now()
		next.ServeHTTP(recorder, r)
		duration := time.Since(start)

		written, _ := strconv.ParseFloat(w.Header().Get("Content-Length"), 64)

		app_serve.HttpRequestsTotal.With(prometheus.Labels{"method": r.Method, "path": r.URL.Path}).Inc()
		app_serve.HttpRequestDuration.With(prometheus.Labels{"method": r.Method, "path": r.URL.Path}).Observe(duration.Seconds())
		app_serve.HttpResponseSizeBytes.With(prometheus.Labels{"method": r.Method, "path": r.URL.Path}).Observe(float64(written))
		app_serve.HttpInFlightRequests.With(prometheus.Labels{"method": r.Method, "path": r.URL.Path}).Dec()
		app_serve.HttpErrorTotal.With(prometheus.Labels{"method": r.Method, "path": r.URL.Path, "status_code": http.StatusText(recorder.Status)}).Inc()
	})
}
