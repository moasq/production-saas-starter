// middleware/request_size.go

package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip check for GET, HEAD, OPTIONS methods
		if c.Request.Method == http.MethodGet ||
			c.Request.Method == http.MethodHead ||
			c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		// Check Content-Length header
		contentLength := c.Request.ContentLength
		if contentLength > maxSize {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": fmt.Sprintf("request size limit exceeded: %d bytes > %d bytes", contentLength, maxSize),
			})
			return
		}

		// Set body size limit for the request
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

		c.Next()

		// Check if the request was aborted due to size limit
		if c.Errors.Last() != nil && c.Errors.Last().Err == http.ErrMissingFile {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "request size limit exceeded during processing",
			})
			return
		}
	}
}
