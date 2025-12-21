package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// Headers
	RequestIDHeader = "X-Request-ID"

	// Context keys
	RequestIDKey = "request_id"
)

// RequestID middleware ensures each request has a unique ID for tracing
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request already has an ID
		requestID := c.GetHeader(RequestIDHeader)

		// Generate new ID if none exists
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Set ID in context and response header
		c.Set(RequestIDKey, requestID)
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}

// generateRequestID creates a new UUID v4 for request tracking
func generateRequestID() string {
	return uuid.New().String()
}

// GetRequestID retrieves request ID from context
func GetRequestID(c *gin.Context) string {
	if id, exists := c.Get(RequestIDKey); exists {
		return id.(string)
	}
	return ""
}
