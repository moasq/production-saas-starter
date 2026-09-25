package cmd

import (
	"github.com/gin-gonic/gin"
	"github.com/moasq/go-b2b-starter/internal/modules/auth"
	"github.com/moasq/go-b2b-starter/internal/platform/server/config"
	"github.com/moasq/go-b2b-starter/internal/platform/server/domain"
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
	container.Provide(func(cfg *config.Config) *gin.Engine {
		if cfg.IsProd() {
			gin.SetMode(gin.ReleaseMode)
		}
		return gin.New()
	})
	container.Provide(domain.NewHTTPServer)

	// Provide server as auth.ServerMiddlewareRegistrar for auth package
	container.Provide(func(srv domain.Server) auth.ServerMiddlewareRegistrar {
		return &serverMiddlewareAdapter{server: srv}
	})

}
