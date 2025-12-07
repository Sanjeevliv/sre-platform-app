package api

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sanjeevsethi/sre-platform-app/internal/metadata"
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
	r.GET("/ready", readyHandler)
	r.GET("/version", versionHandler)
	r.GET("/debug/info", debugInfoHandler)
	// Exposing the /metrics endpoint using the promhttp handler wrapped in gin
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return r
}

// Creating handler for /healthz (Liveness)
func healthzHandler(c *gin.Context) {
	httpRequestTotal.With(prometheus.Labels{"path": "/healthz"}).Inc()
	c.String(http.StatusOK, "ok")
}

// Creating handler for /ready (Readiness)
func readyHandler(c *gin.Context) {
	httpRequestTotal.With(prometheus.Labels{"path": "/ready"}).Inc()
	// Future: Check DB/Redis connections here.
	c.String(http.StatusOK, "ready")
}

// Version Handler
func versionHandler(c *gin.Context) {
	httpRequestTotal.With(prometheus.Labels{"path": "/version"}).Inc()
	c.JSON(http.StatusOK, metadata.GetBuildInfo())
}

// Debug Info Handler
func debugInfoHandler(c *gin.Context) {
	httpRequestTotal.With(prometheus.Labels{"path": "/debug/info"}).Inc()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	c.JSON(http.StatusOK, gin.H{
		"goroutines":         runtime.NumGoroutine(),
		"memory_alloc":       m.Alloc,
		"memory_total_alloc": m.TotalAlloc,
		"memory_sys":         m.Sys,
		"num_gc":             m.NumGC,
	})
}

func rootHandler(c *gin.Context) {
	// Instrument this endpoint call
	httpRequestTotal.With(prometheus.Labels{"path": "/"}).Inc()

	c.String(http.StatusOK, "SRE Platform API Service")
}
