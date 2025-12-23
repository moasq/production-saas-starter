// middleware/recovery.go

package middleware

import (
	"bytes"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/moasq/go-b2b-starter/internal/platform/server/logging"
	"github.com/gin-gonic/gin"
)

const (
	// Size of the stack buffer
	stackSize = 4 << 10 // 4 KB
)

func Recovery(logger *logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get stack trace
				stack := stack(3)

				// Get request details
				httpRequest := c.Request
				headers := make(map[string]string)
				for k, v := range httpRequest.Header {
					headers[k] = v[0]
				}

				// Log the error with context
				logger.Errorw("Panic recovered",
					"error", err,
					"stack", string(stack),
					"request_id", GetRequestID(c),
					"method", httpRequest.Method,
					"path", httpRequest.URL.Path,
					"query", httpRequest.URL.RawQuery,
					"ip", c.ClientIP(),
					"user_agent", httpRequest.UserAgent(),
					"time", time.Now().UTC(),
				)

				// Return safe error to client
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":      "Internal Server Error",
					"request_id": GetRequestID(c),
					"code":       "SERVER_ERROR",
				})
			}
		}()
		c.Next()
	}
}

// stack returns a formatted stack trace of the goroutine that panicked
func stack(_ int) []byte {
	buf := new(bytes.Buffer)

	// Get runtime stack
	var stackBuf [stackSize]byte
	n := runtime.Stack(stackBuf[:], false)

	// Write stack to buffer
	_, _ = io.WriteString(buf, "Stack Trace:\n")
	_, _ = buf.Write(stackBuf[:n])

	return buf.Bytes()
}
