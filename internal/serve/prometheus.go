package serve

const (
	PROMETHEUS_NAMESPACE = "g_fe_server"
	PROMETHEUS_SUBSYSTEM = "presentation_server"
)

// var (
// 	HttpRequestsTotal = prometheus.NewCounterVec(
// 		prometheus.CounterOpts{
// 			Namespace: PROMETHEUS_NAMESPACE,
// 			Subsystem: PROMETHEUS_SUBSYSTEM,
// 			Name:      "http_requests_total",
// 			Help:      "Number of HTTP requests",
// 		},
// 		[]string{"method", "path"},
// 	)
// 	HttpRequestDuration = prometheus.NewHistogramVec(
// 		prometheus.HistogramOpts{
// 			Namespace: PROMETHEUS_NAMESPACE,
// 			Subsystem: PROMETHEUS_SUBSYSTEM,
// 			Name:      "http_request_duration_seconds",
// 			Help:      "Duration of HTTP requests",
// 			Buckets:   prometheus.DefBuckets,
// 		},
// 		[]string{"method", "path"},
// 	)
// 	HttpErrorTotal = prometheus.NewCounterVec(
// 		prometheus.CounterOpts{
// 			Namespace: PROMETHEUS_NAMESPACE,
// 			Subsystem: PROMETHEUS_SUBSYSTEM,
// 			Name:      "http_errors_total",
// 			Help:      "Number of HTTP errors",
// 		},
// 		[]string{"method", "path", "status_code"},
// 	)
// 	HttpInFlightRequests = prometheus.NewGaugeVec(
// 		prometheus.GaugeOpts{
// 			Namespace: PROMETHEUS_NAMESPACE,
// 			Subsystem: PROMETHEUS_SUBSYSTEM,
// 			Name:      "http_in_flight_requests",
// 			Help:      "Number of in-flight HTTP requests",
// 		},
// 		[]string{"method", "path"},
// 	)
// 	HttpResponseSizeBytes = prometheus.NewHistogramVec(
// 		prometheus.HistogramOpts{
// 			Namespace: PROMETHEUS_NAMESPACE,
// 			Subsystem: PROMETHEUS_SUBSYSTEM,
// 			Name:      "http_response_size_bytes",
// 			Help:      "Size of HTTP responses",
// 			Buckets:   prometheus.DefBuckets,
// 		},
// 		[]string{"method", "path"},
// 	)
// )

// func init() {
// 	prometheus.MustRegister(
// 		HttpRequestsTotal,
// 		HttpRequestDuration,
// 		HttpErrorTotal,
// 		HttpInFlightRequests,
// 		HttpResponseSizeBytes,
// 	)
// }
