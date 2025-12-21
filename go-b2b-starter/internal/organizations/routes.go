package organizations

import (
	"github.com/gin-gonic/gin"

	"github.com/moasq/go-b2b-starter/internal/auth"
	serverDomain "github.com/moasq/go-b2b-starter/internal/server/domain"
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
		authGroup.POST("/signup", r.memberHandler.BootstrapOrganization)

		// Public endpoint - Check if email exists (no authentication required)
		authGroup.GET("/check-email", r.memberHandler.CheckEmail)

		// Protected endpoint - Add member (requires JWT authentication)
		authGroup.POST("/members",
			resolver.Get("auth"),
			resolver.Get("org_context"),
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

	// Account routes - require JWT authentication
	accountGroup := router.Group("/accounts")
	accountGroup.Use(
		resolver.Get("auth"),
		resolver.Get("org_context"),
	)
	{
		// Account management
		accountGroup.POST("", auth.RequirePermissionFunc("org", "manage"), r.accountHandler.CreateAccount)
		accountGroup.GET("", auth.RequirePermissionFunc("org", "view"), r.accountHandler.ListAccounts)
		accountGroup.GET("/by-email", auth.RequirePermissionFunc("org", "view"), r.accountHandler.GetAccountByEmail)
		accountGroup.GET("/:id", auth.RequirePermissionFunc("org", "view"), r.accountHandler.GetAccount)
		accountGroup.PUT("/:id", auth.RequirePermissionFunc("org", "manage"), r.accountHandler.UpdateAccount)
		accountGroup.DELETE("/:id", auth.RequirePermissionFunc("org", "manage"), r.accountHandler.DeleteAccount)
		accountGroup.POST("/:id/last-login", auth.RequirePermissionFunc("org", "view"), r.accountHandler.UpdateAccountLastLogin)
		accountGroup.GET("/:id/permissions", auth.RequirePermissionFunc("org", "view"), r.accountHandler.CheckAccountPermission)
		accountGroup.GET("/:id/stats", auth.RequirePermissionFunc("org", "view"), r.accountHandler.GetAccountStats)
	}
}

// Routes returns a RouteRegistrar function compatible with the server interface
func (r *Routes) Routes(router *gin.RouterGroup, resolver serverDomain.MiddlewareResolver) {
	r.RegisterRoutes(router, resolver)
}
