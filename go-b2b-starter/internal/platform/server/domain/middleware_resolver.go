package domain

import "github.com/gin-gonic/gin"

// MiddlewareResolver provides access to named middleware functions
type MiddlewareResolver interface {
	Get(name string) gin.HandlerFunc
}
