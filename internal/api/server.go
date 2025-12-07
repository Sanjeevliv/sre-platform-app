package api

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sanjeevsethi/sre-platform-app/internal/metadata"
)

// NewServer returns a new Gin Engine with all routes registered.
func NewServer(middlewares ...gin.HandlerFunc) *gin.Engine {
	r := gin.New() // Use New() to avoid default Logger/Recovery if we adding our own, or we can add them manually.
	// But sticking to Default() + our own is fine, though double logging might happen if we use ours.
	// The user wanted SRE logs (JSON). Gin default logs to stdout (text).
	// Let's use New() and add Recovery manually. Our logger middleware replaces the default Logger.
	r.Use(gin.Recovery())

	// Add passed middlewares
	for _, m := range middlewares {
		r.Use(m)
	}

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
	c.String(http.StatusOK, "ok")
}

// Creating handler for /ready (Readiness)
func readyHandler(c *gin.Context) {
	// Future: Check DB/Redis connections here.
	c.String(http.StatusOK, "ready")
}

// Version Handler
func versionHandler(c *gin.Context) {
	c.JSON(http.StatusOK, metadata.GetBuildInfo())
}

// Debug Info Handler
func debugInfoHandler(c *gin.Context) {
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
	c.String(http.StatusOK, "SRE Platform API Service")
}
