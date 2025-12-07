package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const RequestIDHeader = "X-Request-ID"

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
