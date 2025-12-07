package api

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Defining my custom Prometheus metric.
// Using a CounterVec to count requests and label them by 'path'.
var (
	httpRequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_service_http_requests_total",
			Help: "Total number of Http requests for the api-service.",
		},
		[]string{"path"},
	)
)

// NewServer returns a new HTTP server multiplexer with all routes registered.
func NewServer() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/healthz", healthzHandler)
	// Exposing the /metrics endpoint using the promhttp handler
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}

// Creating handler for /healthz
func healthzHandler(w http.ResponseWriter, r *http.Request) {
	// Instrument this endpoint call
	httpRequestTotal.With(prometheus.Labels{"path": "/healthz"}).Inc()

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "ok")
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	// Instrument this endpoint call
	httpRequestTotal.With(prometheus.Labels{"path": "/"}).Inc()

	fmt.Fprintln(w, "SRE Platform API Service")
}
