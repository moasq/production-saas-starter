package auth

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Middleware struct {
	provider AuthProvider
	resolver TenantResolver
}

func NewMiddleware(provider AuthProvider, resolver TenantResolver) *Middleware {
	return &Middleware{provider, resolver}
}
func (m *Middleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		cookie := c.GetHeader("Cookie")
		if cookie == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "authentication required"})
			return
		}
		// There is no bearer/JWT alternative that can bypass browser-origin protection.
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" && !sameOrigin(c.Request) {
			c.AbortWithStatusJSON(403, gin.H{"error": "same-origin request required"})
			return
		}
		identity, err := m.provider.VerifySession(c.Request.Context(), cookie)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "authentication required"})
			return
		}
		SetIdentity(c, identity)
		c.Next()
	}
}
func (m *Middleware) RequireOrganization() gin.HandlerFunc {
	return func(c *gin.Context) {
		identity := GetIdentity(c)
		if identity == nil || identity.OrganizationID == "" {
			c.AbortWithStatusJSON(403, gin.H{"error": "organization membership required"})
			return
		}
		ctx, requestContext, err := m.resolver.Resolve(c.Request.Context(), identity)
		if err != nil {
			c.AbortWithStatusJSON(503, gin.H{"error": "workspace unavailable"})
			return
		}
		SetRequestContext(c, requestContext)
		c.Request = c.Request.WithContext(WithRequestContext(ctx, requestContext))
		c.Next()
	}
}
func hasPermission(identity *Identity, resource, action string) bool {
	if identity == nil {
		return false
	}
	required := NewPermission(resource, action)
	for _, permission := range identity.Permissions {
		if permission == required || permission.MatchesWithWildcard(required) {
			return true
		}
	}
	return false
}
func RequirePermissionFunc(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !hasPermission(GetIdentity(c), resource, action) {
			c.AbortWithStatusJSON(403, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}
func sameOrigin(r *http.Request) bool {
	raw := r.Header.Get("Origin")
	origin, err := url.Parse(raw)
	if err != nil || (origin.Scheme != "https" && origin.Scheme != "http") || origin.User != nil || origin.Path != "" || origin.RawQuery != "" || origin.Fragment != "" {
		return false
	}
	if expected := strings.TrimRight(os.Getenv("APP_BASE_URL"), "/"); expected != "" {
		return raw == expected
	}
	return origin.Host == r.Host
}
