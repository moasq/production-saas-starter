package api

import (
	docs "github.com/moasq/go-b2b-starter/internal/docs/gen"
	"github.com/moasq/go-b2b-starter/internal/server/domain"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (h *Handler) Routes(router *gin.RouterGroup, resolver domain.MiddlewareResolver) {
	if gin.Mode() != gin.ReleaseMode {
		docs.SwaggerInfo.Title = "API"
		docs.SwaggerInfo.Description = "API"
		docs.SwaggerInfo.BasePath = "/"

		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}
