package domain

import "github.com/gin-gonic/gin"

// Constants for API versioning
const (
	ApiPrefix   = "/api"
	ApiVersion1 = "v1"
)

// RouteRegistrar is a function type for registering routes to a router group
// domain/server.go
type RouteRegistrar func(*gin.RouterGroup, MiddlewareResolver)

// MiddlewareFunc is a function type that returns a Gin middleware handler
type MiddlewareFunc func() gin.HandlerFunc

// Server defines the interface for HTTP server operations
// domain/server.go - Add to the Server interface
// Server defines the interface for HTTP server operations
type Server interface {
	Start() error
	RegisterRoutes(registrar RouteRegistrar, prefix string, version ...string)
	RegisterNamedMiddleware(name string, middleware MiddlewareFunc)
	MiddlewareResolver() MiddlewareResolver
	GetMiddleware(name string) gin.HandlerFunc // Keep this method for compatibility
}
