package rbac

import (
	"github.com/gin-gonic/gin"

	serverDomain "github.com/moasq/go-b2b-starter/internal/server/domain"
)

// Routes handles RBAC API routes registration
type Routes struct {
	handler *Handler
}

func NewRoutes(handler *Handler) *Routes {
	return &Routes{
		handler: handler,
	}
}

// RegisterRoutes registers RBAC routes on the router
// Note: RBAC endpoints are public and do NOT require authentication
// These endpoints are used by frontend for role/permission discovery
func (r *Routes) RegisterRoutes(router *gin.RouterGroup, resolver serverDomain.MiddlewareResolver) {
	// RBAC info endpoints - NO authentication required for role/permission discovery
	rbacGroup := router.Group("/rbac")
	{
		// Get all roles with their permissions - single source of truth for frontend
		// GET /api/rbac/roles
		rbacGroup.GET("/roles",
			r.handler.GetRoles)

		// Get all permissions - useful for permission checkers
		// GET /api/rbac/permissions
		rbacGroup.GET("/permissions",
			r.handler.GetPermissions)

		// Get permissions organized by category - for structured UI display
		// GET /api/rbac/permissions/by-category
		rbacGroup.GET("/permissions/by-category",
			r.handler.GetPermissionsByCategory)

		// Get detailed information about a specific role with statistics
		// GET /api/rbac/roles/{role_id}
		rbacGroup.GET("/roles/:role_id",
			r.handler.GetRoleDetails)

		// Check if a role has a specific permission - for conditional UI rendering
		// POST /api/rbac/check-permission
		rbacGroup.POST("/check-permission",
			r.handler.CheckPermission)

		// Get RBAC system metadata
		// GET /api/rbac/metadata
		rbacGroup.GET("/metadata",
			r.handler.GetMetadata)
	}
}

// Routes satisfies the RouteRegistrar interface
// This allows the routes to be registered by the server
func (r *Routes) Routes(router *gin.RouterGroup, resolver serverDomain.MiddlewareResolver) {
	r.RegisterRoutes(router, resolver)
}
