// middleware/timeout.go

package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel() // Ensure we call cancel to prevent context leak

		// Create a request with the new context
		c.Request = c.Request.WithContext(ctx)

		// Create a channel to signal completion
		finished := make(chan struct{}, 1)

		// Use a WaitGroup to wait for the goroutine to finish
		var wg sync.WaitGroup
		wg.Add(1)

		// Create a copy of the context writer
		writer := c.Writer

		// Handle the request in a goroutine
		go func() {
			defer wg.Done()

			// Create a wrapped writer to capture the response status
			blw := &bodyLogWriter{ResponseWriter: c.Writer}
			c.Writer = blw

			// Process the handler chain
			c.Next()

			// Signal that the handler is complete
			finished <- struct{}{}
		}()

		// Wait for either completion or timeout
		select {
		case <-finished:
			// Handler completed before timeout
		case <-ctx.Done():
			// Timeout occurred or parent context was cancelled
			if ctx.Err() == context.DeadlineExceeded {
				// Reset the original writer to avoid modifying response after timeout
				c.Writer = writer
				c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{"error": "request timeout"})
			}
		}

		// Wait for the goroutine to finish to prevent potential leaks
		wg.Wait()
	}
}

// bodyLogWriter is a wrapper around ResponseWriter to capture the response status
type bodyLogWriter struct {
	gin.ResponseWriter
	statusCode int
}

func (w *bodyLogWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}
