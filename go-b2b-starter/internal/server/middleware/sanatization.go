// middleware/sanitization.go

package middleware

import (
	"html"
	"net/http"
	"strings"

	"github.com/moasq/go-b2b-starter/internal/server/config"
	"github.com/gin-gonic/gin"
)

func RequestSanitization(config config.SanitizationConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize Path Parameters
		for _, param := range c.Params {
			if containsPathTraversal(param.Value) {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"error": "Invalid path parameter detected",
				})
				return
			}
		}

		// Sanitize Query Parameters
		for _, values := range c.Request.URL.Query() {
			for _, value := range values {
				if !config.DisableXSS && containsXSS(value) {
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
						"error": "Potential XSS detected in query parameter",
					})
					return
				}
				if !config.DisableSQLInjection && containsSQLInjection(value) {
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
						"error": "Potential SQL injection detected in query parameter",
					})
					return
				}
			}
		}

		// Add a custom header indicating security checks passed
		c.Header("X-Content-Security", "sanitized")

		c.Next()
	}
}

func containsPathTraversal(path string) bool {
	suspicious := []string{"..", "//", "\\\\"}
	for _, pattern := range suspicious {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

func containsXSS(input string) bool {
	suspicious := []string{
		"<script>", "</script>",
		"javascript:", "vbscript:",
		"onload=", "onerror=",
	}

	sanitized := html.EscapeString(input)
	for _, pattern := range suspicious {
		if strings.Contains(strings.ToLower(sanitized), strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

func containsSQLInjection(input string) bool {
	suspicious := []string{
		"DROP TABLE",
		"DELETE FROM",
		"INSERT INTO",
		"UPDATE",
		"--",
		"UNION",
		"SELECT",
	}

	for _, pattern := range suspicious {
		if strings.Contains(strings.ToUpper(input), pattern) {
			return true
		}
	}
	return false
}
