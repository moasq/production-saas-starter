package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// Timeout sets a cooperative request deadline, preserving earlier parent deadlines
// and cancellation. Pass c.Request.Context() to database and provider calls.
//
// Handlers and their response writes stay on the request goroutine. This middleware
// neither interrupts handlers nor generates a timeout response: the handler owns
// error translation and must stop work when its context ends. An already-started
// response cannot safely be replaced. The server's WriteTimeout bounds network
// writes; it does not stop a handler that ignores cancellation.
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
