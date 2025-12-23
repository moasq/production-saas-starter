package gin

import (
	"github.com/moasq/go-b2b-starter/internal/platform/server/config"
	"github.com/gin-gonic/gin"
)

type GinRouter struct {
	engine *gin.Engine
	v1     *gin.RouterGroup
}

func NewGinRouter(cfg *config.Config) *GinRouter {
	router := gin.New()
	router.Use(gin.Recovery())
	return &GinRouter{
		engine: router,
		v1:     router.Group("/api/v1"),
	}
}

func (g *GinRouter) GetHandler() *gin.Engine {
	return g.engine
}
