package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// OrganizationResolver looks up organization by provider org ID.
//
// This interface decouples auth middleware from the organizations domain.
// Implement this interface by wrapping your organization repository.
type OrganizationResolver interface {
	// ResolveByProviderID looks up organization by the auth provider's org ID (e.g., Stytch org UUID).
	// Returns the database organization ID (int32) or error if not found.
	ResolveByProviderID(ctx context.Context, providerOrgID string) (int32, error)
}

// AccountResolver looks up account by email within an organization.
//
// This interface decouples auth middleware from the organizations domain.
// Implement this interface by wrapping your account repository.
type AccountResolver interface {
	// ResolveByEmail looks up account by email within the given organization.
	// Returns the database account ID (int32) or error if not found.
	ResolveByEmail(ctx context.Context, orgID int32, email string) (int32, error)
}

// MiddlewareConfig configures the auth middleware behavior.
type MiddlewareConfig struct {
	// ErrorHandler is called when an error occurs. If nil, default JSON responses are used.
	ErrorHandler func(c *gin.Context, statusCode int, message string, err error)
}

// DefaultMiddlewareConfig returns the default middleware configuration.
func DefaultMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		ErrorHandler: defaultErrorHandler,
	}
}

// defaultErrorHandler sends JSON error responses.
func defaultErrorHandler(c *gin.Context, statusCode int, message string, err error) {
	response := gin.H{
		"error":   message,
		"success": false,
	}
	if err != nil && statusCode >= 500 {
		response["detail"] = err.Error()
	}
	c.JSON(statusCode, response)
}

// Middleware provides auth middleware functions.
//
// Use NewMiddleware to create an instance with proper dependencies.
type Middleware struct {
	provider    AuthProvider
	orgResolver OrganizationResolver
	accResolver AccountResolver
	config      *MiddlewareConfig
}

// Parameters:
//   - provider: The auth provider for token verification (e.g., Stytch adapter)
//   - orgResolver: Resolves org by provider ID (optional, required for RequireOrganization)
//   - accResolver: Resolves account by email (optional, required for RequireOrganization)
//   - config: Middleware configuration (optional, uses defaults if nil)
func NewMiddleware(
	provider AuthProvider,
	orgResolver OrganizationResolver,
	accResolver AccountResolver,
	config *MiddlewareConfig,
) *Middleware {
	if config == nil {
		config = DefaultMiddlewareConfig()
	}
	return &Middleware{
		provider:    provider,
		orgResolver: orgResolver,
		accResolver: accResolver,
		config:      config,
	}
}

// RequireAuth returns middleware that verifies the JWT token.
//
// This middleware:
//  1. Extracts Bearer token from Authorization header
//  2. Verifies token using the AuthProvider
//  3. Sets Identity in Gin context (accessible via GetIdentity)
//
// Must be called before any middleware that requires authentication.
//
// Usage:
//
//	router.Use(authMiddleware.RequireAuth())
func (m *Middleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip OPTIONS requests (CORS preflight)
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Extract Bearer token
		token, err := extractBearerToken(c)
		if err != nil {
			m.config.ErrorHandler(c, http.StatusUnauthorized, "missing or invalid authorization header", err)
			c.Abort()
			return
		}

		// Verify token
		identity, err := m.provider.VerifyToken(c.Request.Context(), token)
		if err != nil {
			statusCode := HTTPStatusCode(err)
			message := errorMessage(err)
			m.config.ErrorHandler(c, statusCode, message, err)
			c.Abort()
			return
		}

		// Set identity in context
		SetIdentity(c, identity)

		c.Next()
	}
}

// RequireOrganization returns middleware that resolves org/account from Identity.
//
// This middleware:
//  1. Gets Identity from context (requires RequireAuth to run first)
//  2. Looks up organization by provider org ID
//  3. Looks up account by email within organization
//  4. Sets RequestContext in Gin context (accessible via GetRequestContext)
//
// Must be called after RequireAuth middleware.
//
// Usage:
//
//	router.Use(authMiddleware.RequireAuth())
//	router.Use(authMiddleware.RequireOrganization())
func (m *Middleware) RequireOrganization() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get identity from context
		identity := GetIdentity(c)
		if identity == nil {
			m.config.ErrorHandler(c, http.StatusUnauthorized, "authentication required", nil)
			c.Abort()
			return
		}

		// Validate required fields
		if identity.OrganizationID == "" {
			m.config.ErrorHandler(c, http.StatusForbidden, "no organization in token", ErrMissingOrganization)
			c.Abort()
			return
		}

		if identity.Email == "" {
			m.config.ErrorHandler(c, http.StatusForbidden, "no email in token", ErrMissingEmail)
			c.Abort()
			return
		}

		// Resolve organization
		orgID, err := m.orgResolver.ResolveByProviderID(c.Request.Context(), identity.OrganizationID)
		if err != nil {
			m.config.ErrorHandler(c, http.StatusForbidden, "organization not found", err)
			c.Abort()
			return
		}

		// Resolve account
		accountID, err := m.accResolver.ResolveByEmail(c.Request.Context(), orgID, identity.Email)
		if err != nil {
			m.config.ErrorHandler(c, http.StatusForbidden, "account not found", err)
			c.Abort()
			return
		}

		// Set request context
		reqCtx := &RequestContext{
			Identity:       identity,
			OrganizationID: orgID,
			AccountID:      accountID,
			ProviderOrgID:  identity.OrganizationID,
		}
		SetRequestContext(c, reqCtx)

		// Also set individual values for backward compatibility
		c.Set("organization_id", orgID)
		c.Set("account_id", accountID)
		c.Set("stytch_org_id", identity.OrganizationID)

		c.Next()
	}
}

// RequirePermission returns middleware that checks for a specific permission.
//
// This middleware:
//  1. Gets Identity from context (requires RequireAuth to run first)
//  2. Checks if user has the required permission
//  3. Falls back to role-based permissions if not found in Identity
//
// Must be called after RequireAuth middleware.
//
// Usage:
//
//	router.GET("/invoices", authMiddleware.RequirePermission("invoice", "view"), handler)
//	router.POST("/invoices", authMiddleware.RequirePermission("invoice", "create"), handler)
func (m *Middleware) RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		identity := GetIdentity(c)
		if identity == nil {
			m.config.ErrorHandler(c, http.StatusUnauthorized, "authentication required", nil)
			c.Abort()
			return
		}

		if !hasPermission(identity, resource, action) {
			m.config.ErrorHandler(c, http.StatusForbidden, "insufficient permissions", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission returns middleware that checks for any of the given permissions.
//
// This middleware succeeds if the user has at least one of the specified permissions.
// Useful when multiple permissions can grant access to the same resource.
//
// Must be called after RequireAuth middleware.
//
// Usage:
//
//	router.GET("/reports", authMiddleware.RequireAnyPermission(
//	    auth.PermPaymentOptSchedule,
//	    auth.PermPaymentOptExport,
//	), handler)
func (m *Middleware) RequireAnyPermission(permissions ...Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		identity := GetIdentity(c)
		if identity == nil {
			m.config.ErrorHandler(c, http.StatusUnauthorized, "authentication required", nil)
			c.Abort()
			return
		}

		for _, perm := range permissions {
			if hasPermission(identity, perm.Resource(), perm.Action()) {
				c.Next()
				return
			}
		}

		m.config.ErrorHandler(c, http.StatusForbidden, "insufficient permissions", nil)
		c.Abort()
	}
}

// RequireAllPermissions returns middleware that checks for all given permissions.
//
// This middleware succeeds only if the user has all of the specified permissions.
//
// Must be called after RequireAuth middleware.
//
// Usage:
//
//	router.DELETE("/org", authMiddleware.RequireAllPermissions(
//	    auth.NewPermission("org", "view"),
//	    auth.NewPermission("org", "manage"),
//	), handler)
func (m *Middleware) RequireAllPermissions(permissions ...Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		identity := GetIdentity(c)
		if identity == nil {
			m.config.ErrorHandler(c, http.StatusUnauthorized, "authentication required", nil)
			c.Abort()
			return
		}

		for _, perm := range permissions {
			if !hasPermission(identity, perm.Resource(), perm.Action()) {
				m.config.ErrorHandler(c, http.StatusForbidden, "insufficient permissions", nil)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// RequireRole returns middleware that checks for a specific role.
//
// Must be called after RequireAuth middleware.
//
// Usage:
//
//	router.POST("/admin", authMiddleware.RequireRole(auth.RoleAdmin), handler)
func (m *Middleware) RequireRole(role Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		identity := GetIdentity(c)
		if identity == nil {
			m.config.ErrorHandler(c, http.StatusUnauthorized, "authentication required", nil)
			c.Abort()
			return
		}

		if !hasRole(identity, role) {
			m.config.ErrorHandler(c, http.StatusForbidden, "insufficient role", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole returns middleware that checks for any of the given roles.
//
// Must be called after RequireAuth middleware.
//
// Usage:
//
//	router.GET("/dashboard", authMiddleware.RequireAnyRole(auth.RoleAdmin, auth.RoleApprover), handler)
func (m *Middleware) RequireAnyRole(roles ...Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		identity := GetIdentity(c)
		if identity == nil {
			m.config.ErrorHandler(c, http.StatusUnauthorized, "authentication required", nil)
			c.Abort()
			return
		}

		for _, role := range roles {
			if hasRole(identity, role) {
				c.Next()
				return
			}
		}

		m.config.ErrorHandler(c, http.StatusForbidden, "insufficient role", nil)
		c.Abort()
	}
}

// Helper functions

// extractBearerToken extracts the JWT from Authorization header.
func extractBearerToken(c *gin.Context) (string, error) {
	header := c.GetHeader("Authorization")
	if header == "" {
		return "", ErrUnauthorized
	}

	fields := strings.Fields(header)
	if len(fields) != 2 || !strings.EqualFold(fields[0], "bearer") {
		return "", ErrInvalidToken
	}

	return fields[1], nil
}

// hasPermission checks if identity has the required permission.
func hasPermission(identity *Identity, resource, action string) bool {
	perm := NewPermission(resource, action)

	// Check explicit permissions in identity
	for _, p := range identity.Permissions {
		if p == perm || p.MatchesWithWildcard(perm) {
			return true
		}
	}

	// Fallback: Check role-based permissions
	for _, role := range identity.Roles {
		if HasRolePermission(role, resource, action) {
			return true
		}
	}

	return false
}

// hasRole checks if identity has the required role.
func hasRole(identity *Identity, role Role) bool {
	normalized := NormalizeRole(string(role))
	for _, r := range identity.Roles {
		if NormalizeRole(string(r)) == normalized {
			return true
		}
	}
	return false
}

// errorMessage returns a user-friendly message for auth errors.
func errorMessage(err error) string {
	switch err {
	case ErrTokenExpired:
		return "token expired"
	case ErrInvalidToken:
		return "invalid token"
	case ErrEmailNotVerified:
		return "email not verified"
	case ErrAudienceMismatch:
		return "invalid token audience"
	case ErrIssuerMismatch:
		return "invalid token issuer"
	default:
		return "authentication failed"
	}
}

// Standalone middleware functions for simpler usage

// RequirePermissionFunc returns a standalone middleware that checks permissions.
//
// This is a convenience function that doesn't require a Middleware instance.
// It reads Identity directly from Gin context.
//
// Usage:
//
//	router.GET("/invoices", auth.RequirePermissionFunc("invoice", "view"), handler)
func RequirePermissionFunc(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		identity := GetIdentity(c)
		if identity == nil {
			defaultErrorHandler(c, http.StatusUnauthorized, "authentication required", nil)
			c.Abort()
			return
		}

		if !hasPermission(identity, resource, action) {
			defaultErrorHandler(c, http.StatusForbidden, "insufficient permissions", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermissionFunc returns a standalone middleware that checks for any permission.
//
// Usage:
//
//	router.GET("/reports", auth.RequireAnyPermissionFunc(
//	    auth.PermPaymentOptSchedule,
//	    auth.PermPaymentOptExport,
//	), handler)
func RequireAnyPermissionFunc(permissions ...Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		identity := GetIdentity(c)
		if identity == nil {
			defaultErrorHandler(c, http.StatusUnauthorized, "authentication required", nil)
			c.Abort()
			return
		}

		for _, perm := range permissions {
			if hasPermission(identity, perm.Resource(), perm.Action()) {
				c.Next()
				return
			}
		}

		defaultErrorHandler(c, http.StatusForbidden, "insufficient permissions", nil)
		c.Abort()
	}
}
