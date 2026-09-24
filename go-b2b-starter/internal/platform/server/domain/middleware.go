package domain

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/moasq/go-b2b-starter/internal/platform/logger"
	"github.com/moasq/go-b2b-starter/internal/platform/server/middleware"
)

func (s *HTTPServer) setupMiddleware() {
	requestTimeout := 25 * time.Second

	s.router.Use(
		middleware.RequestID(),
		middleware.Recovery(s.logger),
		middleware.RequestSizeLimit(int64(s.config.MaxRequestSize)),
		middleware.Timeout(requestTimeout),
		middleware.RateLimiter(s.config.RateLimitPerSecond),
		middleware.CORS(s.config.AllowedOrigins),
		s.requestLoggingMiddleware(),
	)

	// production only middleware
	if s.config.IsProd() {
		s.router.Use(
			middleware.SecurityHeaders(),
		)
	}

	if err := s.router.SetTrustedProxies(s.config.TrustedProxies); err != nil {
		panic(err)
	}
}

func (s *HTTPServer) requestLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip health check logging in production
		if s.config.IsProd() && c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		requestID := middleware.GetRequestID(c) // Get request ID

		c.Next()

		s.logger.Info("Request completed", logger.Fields{
			"request_id": requestID,
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       path,
			"ip":         c.ClientIP(),
			"latency_ms": time.Since(start).Milliseconds(),
			"bytes_out":  c.Writer.Size(),
		})
	}
}
