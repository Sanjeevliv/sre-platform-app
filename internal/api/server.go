package api

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/sanjeevsethi/sre-platform-app/internal/metadata"
	"github.com/sanjeevsethi/sre-platform-app/internal/queue"
)

// NewServer returns a new Gin Engine with all routes registered.
func NewServer(producer *queue.Producer, middlewares ...gin.HandlerFunc) *gin.Engine {
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

	// Jobs endpoint
	r.POST("/jobs", func(c *gin.Context) {
		jobHandler(c, producer)
	})

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

type JobRequest struct {
	Payload string `json:"payload"`
}

func jobHandler(c *gin.Context, p *queue.Producer) {
	var req JobRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	rid := c.GetString("request_id")
	if rid == "" {
		rid = "unknown"
	}

	job := queue.Job{
		ID:        uuid.New().String(),
		Payload:   req.Payload,
		RequestID: rid,
	}

	ctx := c.Request.Context()
	if err := p.Enqueue(ctx, job); err != nil {
		// Circuit breaker error or Redis error
		log.Error().Err(err).Msg("Failed to enqueue job")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service unavailable"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "queued", "job_id": job.ID})
}
