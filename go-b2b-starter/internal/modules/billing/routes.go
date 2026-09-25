package billing

import (
	"github.com/gin-gonic/gin"
	"github.com/moasq/go-b2b-starter/internal/modules/auth"
	serverDomain "github.com/moasq/go-b2b-starter/internal/platform/server/domain"
)

func (h *Handler) Routes(router *gin.RouterGroup, resolver serverDomain.MiddlewareResolver) {
	group := router.Group("/subscriptions", resolver.Get("auth"), resolver.Get("org_context"))
	group.GET("/status", auth.RequirePermissionFunc("org", "view"), h.GetBillingStatus)
	group.POST("/verify-payment", auth.RequirePermissionFunc("org", "manage"), h.VerifyPayment)
}
