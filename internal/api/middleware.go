package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/time/rate"
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

// RateLimitMiddleware creates a token bucket rate limiter.
func RateLimitMiddleware(rps int, burst int) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(rps), burst)
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests",
			})
			return
		}
		c.Next()
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

		// Extract trace_id and span_id from OpenTelemetry context
		var traceID, spanID string
		span := trace.SpanFromContext(c.Request.Context())
		if span.SpanContext().IsValid() {
			traceID = span.SpanContext().TraceID().String()
			spanID = span.SpanContext().SpanID().String()
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
			Str("trace_id", traceID).
			Str("span_id", spanID).
			Msg("Request processed")
	}
}
