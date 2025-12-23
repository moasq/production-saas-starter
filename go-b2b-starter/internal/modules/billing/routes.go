package billing

import (
	"github.com/gin-gonic/gin"

	"github.com/moasq/go-b2b-starter/internal/modules/auth"
	serverDomain "github.com/moasq/go-b2b-starter/internal/platform/server/domain"
)

// Routes registers subscription endpoints
func (h *Handler) Routes(router *gin.RouterGroup, resolver serverDomain.MiddlewareResolver) {
	// Subscription endpoints
	subscriptions := router.Group("/subscriptions")
	subscriptions.Use(
		resolver.Get("auth"),
		resolver.Get("org_context"),
	)
	{
		// Get billing status - requires resource:view permission
		subscriptions.GET("/status",
			auth.RequirePermissionFunc("resource", "view"),
			h.GetBillingStatus)
	}

	// Verify payment endpoint - auth only (session_id identifies org)
	// This is separate from the main group to avoid requiring org_context middleware
	// The session_id from the checkout contains the customer_id which maps to the org
	router.POST("/subscriptions/verify-payment",
		resolver.Get("auth"),
		h.VerifyPayment)
}
