package middleware

import (
	"github.com/gin-gonic/gin"
)

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-no-referrer")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; img-src 'self' https:; style-src 'self' 'unsafe-inline';")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("X-Permitted-Cross-Domain-Policies", "none")

		// Remove sensitive headers
		c.Header("Server", "")
		c.Header("X-Powered-By", "")

		c.Next()
	}
}
