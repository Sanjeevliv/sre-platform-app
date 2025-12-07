package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

// NewServer returns a new Gin Engine with all routes registered.
func NewServer() *gin.Engine {
	r := gin.Default()

	r.GET("/", rootHandler)
	r.GET("/healthz", healthzHandler)
	// Exposing the /metrics endpoint using the promhttp handler wrapped in gin
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return r
}

// Creating handler for /healthz
func healthzHandler(c *gin.Context) {
	// Instrument this endpoint call
	httpRequestTotal.With(prometheus.Labels{"path": "/healthz"}).Inc()

	c.String(http.StatusOK, "ok")
}

func rootHandler(c *gin.Context) {
	// Instrument this endpoint call
	httpRequestTotal.With(prometheus.Labels{"path": "/"}).Inc()

	c.String(http.StatusOK, "SRE Platform API Service")
}
