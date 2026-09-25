package organizations

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/moasq/go-b2b-starter/internal/modules/auth"
	"github.com/moasq/go-b2b-starter/internal/platform/betterauth"
	server "github.com/moasq/go-b2b-starter/internal/platform/server/domain"
)

type Routes struct{ bridge *betterauth.Client }

func NewRoutes(bridge *betterauth.Client) *Routes { return &Routes{bridge} }
func (r *Routes) Routes(router *gin.RouterGroup, resolver server.MiddlewareResolver) {
	group := router.Group("/auth", resolver.Get("auth"), resolver.Get("org_context"))
	group.GET("/profile/me", r.profile)
	group.PUT("/profile/me", r.updateProfile)
	members := group.Group("/members", auth.RequirePermissionFunc("org", "manage"))
	members.GET("", func(c *gin.Context) { r.forward(c, "members-list", struct{}{}) })
	members.POST("", r.invite)
	members.DELETE("/:member_id", func(c *gin.Context) { r.forward(c, "members-remove", gin.H{"member_id": c.Param("member_id")}) })
	members.POST("/:member_id/resend-invitation", func(c *gin.Context) { r.forward(c, "members-resend", gin.H{"member_id": c.Param("member_id")}) })
	members.PUT("/:member_id", r.updateRole)
	org := router.Group("/organizations", resolver.Get("auth"), resolver.Get("org_context"))
	org.GET("", auth.RequirePermissionFunc("org", "view"), func(c *gin.Context) {
		i := auth.GetIdentity(c)
		c.JSON(200, gin.H{"success": true, "data": i.Organization})
	})
	org.PUT("", auth.RequirePermissionFunc("org", "manage"), r.updateOrganization)
}
func (r *Routes) profile(c *gin.Context) {
	scope := auth.GetRequestContext(c)
	i := scope.Identity
	c.JSON(200, gin.H{"success": true, "data": gin.H{
		"member_id": i.MemberID, "email": i.Email, "name": i.User.Name, "roles": i.Roles, "permissions": i.Permissions, "email_verified": i.EmailVerified, "status": "active",
		"organization": gin.H{"organization_id": i.OrganizationID, "slug": i.Organization.Slug, "name": i.Organization.Name, "status": "active"},
		"account_id":   scope.AccountID, "created_at": scope.CreatedAt, "updated_at": scope.UpdatedAt,
	}})
}
func (r *Routes) updateProfile(c *gin.Context) {
	var body struct {
		Name string `json:"name" binding:"required,min=1,max=255"`
	}
	if c.ShouldBindJSON(&body) != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "valid name required"})
		return
	}
	r.forward(c, "profile", body)
}
func (r *Routes) updateOrganization(c *gin.Context) {
	var body struct {
		Name string `json:"name" binding:"required,min=1,max=255"`
	}
	if c.ShouldBindJSON(&body) != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "valid name required"})
		return
	}
	r.forward(c, "organization-update", body)
}
func (r *Routes) invite(c *gin.Context) {
	var body struct {
		Email string `json:"email" binding:"required,email,max=255"`
		Name  string `json:"name" binding:"required,min=1,max=255"`
		Role  string `json:"role_slug"`
	}
	if c.ShouldBindJSON(&body) != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "valid email and name required"})
		return
	}
	if body.Role == "" {
		body.Role = "member"
	}
	if !validRole(body.Role) {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid role"})
		return
	}
	r.forward(c, "members-invite", gin.H{"email": body.Email, "name": body.Name, "role": body.Role})
}
func (r *Routes) updateRole(c *gin.Context) {
	var body struct {
		Role string `json:"role"`
	}
	if c.ShouldBindJSON(&body) != nil || !validRole(body.Role) {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid role"})
		return
	}
	r.forward(c, "members-role", gin.H{"member_id": c.Param("member_id"), "role": body.Role})
}
func validRole(role string) bool { return role == "admin" || role == "manager" || role == "member" }
func (r *Routes) forward(c *gin.Context, operation string, body any) {
	status, data, err := r.bridge.Call(c.Request.Context(), c.GetHeader("Cookie"), operation, body)
	if err != nil || status >= 500 || status < 200 || status >= 400 && status != 400 && status != 401 && status != 403 && status != 404 && status != 409 && status != 429 || !json.Valid(data) {
		c.AbortWithStatusJSON(502, gin.H{"error": "authentication service unavailable"})
		return
	}
	c.Data(status, "application/json", data)
}
