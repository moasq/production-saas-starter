package auth

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

type ServerMiddlewareRegistrar interface {
	RegisterNamedMiddleware(string, func() gin.HandlerFunc)
}

func RegisterNamedMiddlewares(container *dig.Container) error {
	return container.Invoke(func(m *Middleware, s ServerMiddlewareRegistrar) {
		s.RegisterNamedMiddleware("auth", m.RequireAuth)
		s.RegisterNamedMiddleware("org_context", m.RequireOrganization)
	})
}
