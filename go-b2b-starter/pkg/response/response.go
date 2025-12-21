package response

import (
	"github.com/gin-gonic/gin"

	"github.com/moasq/go-b2b-starter/pkg/httperr"
)

// Success sends a successful response
func Success(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, gin.H{
		"success": true,
		"data":    data,
	})
}

// Error sends an error response
func Error(c *gin.Context, statusCode int, message string, err error) {
	c.JSON(statusCode, errors.NewHTTPError(
		statusCode,
		"error",
		message,
	))
}