package organizations

import (
	"github.com/gin-gonic/gin"
	platformstytch "github.com/moasq/go-b2b-starter/internal/platform/stytch"

	"github.com/moasq/go-b2b-starter/internal/modules/auth"
	serverDomain "github.com/moasq/go-b2b-starter/internal/platform/server/domain"
)

type Routes struct {
	organizationHandler *OrganizationHandler
	accountHandler      *AccountHandler
	memberHandler       *MemberHandler
}

func NewRoutes(
	organizationHandler *OrganizationHandler,
	accountHandler *AccountHandler,
	memberHandler *MemberHandler,
) *Routes {
	return &Routes{
		organizationHandler: organizationHandler,
		accountHandler:      accountHandler,
		memberHandler:       memberHandler,
	}
}

// RegisterRoutes registers organization, account, and auth member management routes
func (r *Routes) RegisterRoutes(router *gin.RouterGroup, resolver serverDomain.MiddlewareResolver) {
	// Auth routes - member management and authentication
	authGroup := router.Group("/auth")
	{
		// Public endpoint - Organization signup (no authentication required)
		authGroup.POST("/signup", requireConfiguredAuth(), r.memberHandler.BootstrapOrganization)

		// Protected endpoint - Add member (requires JWT authentication)
		authGroup.POST("/members",
			resolver.Get("auth"),
			resolver.Get("org_context"),
			auth.RequirePermissionFunc("org", "manage"),
			r.memberHandler.AddMember)

		// Protected endpoint - List members (requires JWT authentication and org:manage permission)
		authGroup.GET("/members",
			resolver.Get("auth"),
			resolver.Get("org_context"),
			auth.RequirePermissionFunc("org", "manage"),
			r.memberHandler.ListMembers)

		// Protected endpoint - Get current user profile (requires JWT authentication only)
		authGroup.GET("/profile/me",
			resolver.Get("auth"),
			resolver.Get("org_context"),
			r.memberHandler.GetProfile)

		authGroup.PUT("/profile/me", resolver.Get("auth"), resolver.Get("org_context"), r.memberHandler.UpdateProfile)
		authGroup.POST("/members/:member_id/resend-invitation", resolver.Get("auth"), resolver.Get("org_context"), auth.RequirePermissionFunc("org", "manage"), r.memberHandler.ResendInvitation)

		// Protected endpoint - Delete organization member (requires JWT authentication and org:manage permission)
		authGroup.DELETE("/members/:member_id",
			resolver.Get("auth"),
			resolver.Get("org_context"),
			auth.RequirePermissionFunc("org", "manage"),
			r.memberHandler.DeleteMember)
	}

	// Organization routes - require JWT authentication
	orgGroup := router.Group("/organizations")
	orgGroup.Use(
		resolver.Get("auth"),
		resolver.Get("org_context"),
	)
	{
		// Current organization endpoints
		orgGroup.GET("", auth.RequirePermissionFunc("org", "view"), r.organizationHandler.GetOrganization)
		orgGroup.PUT("", auth.RequirePermissionFunc("org", "manage"), r.organizationHandler.UpdateOrganization)
		orgGroup.GET("/stats", auth.RequirePermissionFunc("org", "view"), r.organizationHandler.GetOrganizationStats)
	}

}

// Routes returns a RouteRegistrar function compatible with the server interface
func (r *Routes) Routes(router *gin.RouterGroup, resolver serverDomain.MiddlewareResolver) {
	r.RegisterRoutes(router, resolver)
}

func requireConfiguredAuth() gin.HandlerFunc {
	cfg, err := platformstytch.LoadConfig()
	return func(c *gin.Context) {
		if err != nil || !cfg.Configured() {
			c.AbortWithStatusJSON(503, gin.H{"error": "Authentication is not configured"})
			return
		}
		c.Next()
	}
}
