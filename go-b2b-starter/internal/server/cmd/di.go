package cmd

import (
	"github.com/gin-gonic/gin"
	"github.com/moasq/go-b2b-starter/internal/auth"
	"github.com/moasq/go-b2b-starter/internal/paywall"
	"github.com/moasq/go-b2b-starter/internal/server/config"
	"github.com/moasq/go-b2b-starter/internal/server/domain"
	ginP "github.com/moasq/go-b2b-starter/internal/server/gin"
	"github.com/moasq/go-b2b-starter/internal/server/logging"
	"github.com/moasq/go-b2b-starter/internal/server/middleware"
	"go.uber.org/dig"
)

// serverMiddlewareAdapter adapts domain.Server to auth.ServerMiddlewareRegistrar
type serverMiddlewareAdapter struct {
	server domain.Server
}

func (a *serverMiddlewareAdapter) RegisterNamedMiddleware(name string, middleware func() gin.HandlerFunc) {
	// Convert func() gin.HandlerFunc to domain.MiddlewareFunc
	a.server.RegisterNamedMiddleware(name, domain.MiddlewareFunc(middleware))
}

func SetupDependencies(container *dig.Container) {
	container.Provide(config.LoadConfig)
	container.Provide(logging.InitLogger)
	container.Provide(middleware.InitValidator)
	container.Provide(func(cfg *config.Config) *gin.Engine {
		return ginP.NewGinRouter(cfg).GetHandler()
	})
	container.Provide(domain.NewHTTPServer)

	// Provide server as auth.ServerMiddlewareRegistrar for auth package
	container.Provide(func(srv domain.Server) auth.ServerMiddlewareRegistrar {
		return &serverMiddlewareAdapter{server: srv}
	})

	// Provide server as paywall.ServerMiddlewareRegistrar for paywall package
	container.Provide(func(srv domain.Server) paywall.ServerMiddlewareRegistrar {
		return &serverMiddlewareAdapter{server: srv}
	})
}
