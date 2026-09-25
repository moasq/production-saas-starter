package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"time"
)

// Timeout propagates cancellation to database and provider calls. Handler execution
// stays on the Gin request goroutine; concurrent writes to its context are unsafe.
// The HTTP server's write timeout bounds stalled clients.
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
