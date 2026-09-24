package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/moasq/go-b2b-starter/internal/platform/logger"
)

func Recovery(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Do not log query strings, headers or bodies: they may contain credentials.
				log.Error("Panic recovered", logger.Fields{
					"error":      err,
					"stack":      string(debug.Stack()),
					"request_id": GetRequestID(c),
					"method":     c.Request.Method,
					"path":       c.Request.URL.Path,
				})
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
