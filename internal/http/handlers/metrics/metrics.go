package metrics

import (
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func PrometheusHandlers(nonFunctionalRouter *mux.Router, ctxRoot string) {

	nonFunctionalRouter.Handle("/metrics", promhttp.Handler()).Name(ctxRoot + "/g/metrics")
}
