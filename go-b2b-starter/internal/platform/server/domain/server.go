package domain

import "github.com/gin-gonic/gin"

const ApiPrefix = "/api"

// RouteRegistrar is a function type for registering routes to a router group
type RouteRegistrar func(*gin.RouterGroup, MiddlewareResolver)

// MiddlewareFunc is a function type that returns a Gin middleware handler
type MiddlewareFunc func() gin.HandlerFunc

// Server defines the interface for HTTP server operations
type Server interface {
	Start() error
	RegisterRoutes(registrar RouteRegistrar, prefix string, version ...string)
	RegisterNamedMiddleware(name string, middleware MiddlewareFunc)
}
