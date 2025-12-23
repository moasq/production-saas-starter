package auth

import (
	"context"

	"github.com/gin-gonic/gin"
)

// Context keys for storing auth data.
// Using unexported type to prevent collisions with other packages.
type contextKey string

const (
	// identityKey is the context key for storing the authenticated Identity.
	identityKey contextKey = "auth_identity"

	// requestContextKey is the context key for storing the RequestContext.
	requestContextKey contextKey = "auth_request_context"
)

// SetIdentity stores the Identity in the Gin context.
//
// This is called by the RequireAuth middleware after successful authentication.
// Application code should not call this directly.
func SetIdentity(c *gin.Context, identity *Identity) {
	c.Set(string(identityKey), identity)
}

// GetIdentity retrieves the Identity from the Gin context.
//
// Returns nil if no identity is set (user not authenticated).
// Use MustGetIdentity if you expect authentication middleware to have run.
//
// Example:
//
//	identity := auth.GetIdentity(c)
//	if identity == nil {
//	    // Handle unauthenticated request
//	}
func GetIdentity(c *gin.Context) *Identity {
	if val, exists := c.Get(string(identityKey)); exists {
		if identity, ok := val.(*Identity); ok {
			return identity
		}
	}
	return nil
}

// MustGetIdentity retrieves the Identity from the Gin context.
//
// Panics if no identity is set. Only use this after RequireAuth middleware.
// For handlers where authentication is optional, use GetIdentity instead.
func MustGetIdentity(c *gin.Context) *Identity {
	identity := GetIdentity(c)
	if identity == nil {
		panic("auth: MustGetIdentity called without Identity in context - ensure RequireAuth middleware is applied")
	}
	return identity
}

// SetRequestContext stores the RequestContext in the Gin context.
//
// This is called by the RequireOrganization middleware after resolving
// the database IDs. Application code should not call this directly.
func SetRequestContext(c *gin.Context, reqCtx *RequestContext) {
	c.Set(string(requestContextKey), reqCtx)
}

// GetRequestContext retrieves the RequestContext from the Gin context.
//
// Returns nil if no request context is set.
// Use MustGetRequestContext if you expect the organization middleware to have run.
//
// Example:
//
//	reqCtx := auth.GetRequestContext(c)
//	if reqCtx == nil {
//	    // Handle request without organization context
//	}
//	orgID := reqCtx.OrganizationID  // int32, type-safe
func GetRequestContext(c *gin.Context) *RequestContext {
	if val, exists := c.Get(string(requestContextKey)); exists {
		if reqCtx, ok := val.(*RequestContext); ok {
			return reqCtx
		}
	}
	return nil
}

// MustGetRequestContext retrieves the RequestContext from the Gin context.
//
// Panics if no request context is set. Only use this after RequireOrganization middleware.
// For handlers where organization context is optional, use GetRequestContext instead.
//
// Example:
//
//	reqCtx := auth.MustGetRequestContext(c)
//	orgID := reqCtx.OrganizationID
//	accountID := reqCtx.AccountID
func MustGetRequestContext(c *gin.Context) *RequestContext {
	reqCtx := GetRequestContext(c)
	if reqCtx == nil {
		panic("auth: MustGetRequestContext called without RequestContext in context - ensure RequireOrganization middleware is applied")
	}
	return reqCtx
}

// GetOrganizationID is a convenience function to get the database organization ID.
//
// Returns 0 if no request context is set.
// Use MustGetRequestContext().OrganizationID if you expect the middleware to have run.
func GetOrganizationID(c *gin.Context) int32 {
	if reqCtx := GetRequestContext(c); reqCtx != nil {
		return reqCtx.OrganizationID
	}
	return 0
}

// GetAccountID is a convenience function to get the database account ID.
//
// Returns 0 if no request context is set.
// Use MustGetRequestContext().AccountID if you expect the middleware to have run.
func GetAccountID(c *gin.Context) int32 {
	if reqCtx := GetRequestContext(c); reqCtx != nil {
		return reqCtx.AccountID
	}
	return 0
}

// WithIdentity adds the Identity to a context.Context.
//
// This is useful for passing auth context through service layers
// that don't use Gin context directly.
func WithIdentity(ctx context.Context, identity *Identity) context.Context {
	return context.WithValue(ctx, identityKey, identity)
}

// IdentityFromContext retrieves the Identity from a context.Context.
//
// Returns nil if no identity is set.
func IdentityFromContext(ctx context.Context) *Identity {
	if val := ctx.Value(identityKey); val != nil {
		if identity, ok := val.(*Identity); ok {
			return identity
		}
	}
	return nil
}

// WithRequestContext adds the RequestContext to a context.Context.
//
// This is useful for passing auth context through service layers
// that don't use Gin context directly.
func WithRequestContext(ctx context.Context, reqCtx *RequestContext) context.Context {
	return context.WithValue(ctx, requestContextKey, reqCtx)
}

// RequestContextFromContext retrieves the RequestContext from a context.Context.
//
// Returns nil if no request context is set.
func RequestContextFromContext(ctx context.Context) *RequestContext {
	if val := ctx.Value(requestContextKey); val != nil {
		if reqCtx, ok := val.(*RequestContext); ok {
			return reqCtx
		}
	}
	return nil
}
