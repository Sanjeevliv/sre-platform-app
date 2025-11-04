package main

import (
	"fmt"
	"log"
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

func main() {
	// Register your handlers
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/healthz", healthzHandler)

	// Exposing the /metrics endpoint using the promhttp handler
	http.Handle("/metrics", promhttp.Handler())

	log.Println("Starting api_service on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
