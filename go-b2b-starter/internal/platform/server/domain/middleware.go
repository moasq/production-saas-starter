package domain

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/moasq/go-b2b-starter/internal/platform/server/middleware"
)

func (s *HTTPServer) setupMiddleware() {
	ipProtection := middleware.NewIPProtection()

	requestTimeout := 25 * time.Second

	s.router.Use(
		middleware.RequestID(),
		ipProtection.Protect(),
		middleware.RequestSanitization(s.config.GetSanitizationConfig()),
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
		query := c.Request.URL.RawQuery
		requestID := middleware.GetRequestID(c) // Get request ID

		c.Next()

		s.logger.Infow("Request completed",
			"request_id", requestID,
			"status", c.Writer.Status(),
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"ip", c.ClientIP(),
			"latency", time.Since(start),
			"user-agent", c.Request.UserAgent(),
			"bytes-out", c.Writer.Size(),
		)
	}
}
