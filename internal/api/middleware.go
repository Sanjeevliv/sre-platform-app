package api

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
)

const RequestIDHeader = "X-Request-ID"

// Metrics
var (
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_http_request_duration_seconds",
			Help:    "Duration of HTTP requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
)

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for incoming header
		rid := c.GetHeader(RequestIDHeader)
		if rid == "" {
			rid = uuid.New().String()
		}

		// Set header for response
		c.Header(RequestIDHeader, rid)

		// Set in context for handlers to use
		c.Set("request_id", rid)

		c.Next()
	}
}

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		path := c.Request.URL.Path

		httpRequestDuration.WithLabelValues(c.Request.Method, path, status).Observe(duration)
	}
}

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log after request processing
		rid := c.GetString("request_id")
		duration := time.Since(start)

		if raw != "" {
			path = path + "?" + raw
		}

		logger := log.Info()
		if c.Writer.Status() >= 500 {
			logger = log.Error()
		} else if c.Writer.Status() >= 400 {
			logger = log.Warn()
		}

		logger.
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", c.Writer.Status()).
			Dur("duration", duration).
			Str("client_ip", c.ClientIP()).
			Str("request_id", rid).
			Msg("Request processed")
	}
}
