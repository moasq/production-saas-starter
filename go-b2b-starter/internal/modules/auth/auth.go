// Package auth provides a unified authentication and authorization layer.
//
// This package abstracts away the authentication provider (Stytch, Auth0, etc.)
// and provides a clean interface for the rest of the application to use.
//
// # Architecture
//
// The auth package follows the adapter pattern:
//
//	┌─────────────────────────────────────────────────────────────────┐
//	│                        Application Layer                        │
//	│  (handlers, services - use auth.GetRequestContext, auth.RequirePermission) │
//	└─────────────────────────────────────────────────────────────────┘
//	                              │
//	                              ▼
//	┌─────────────────────────────────────────────────────────────────┐
//	│                         auth package                            │
//	│  • AuthProvider interface                                       │
//	│  • Identity (provider-agnostic user representation)            │
//	│  • RequestContext (resolved database IDs)                      │
//	│  • Middleware (RequireAuth, RequireOrganization, RequirePermission) │
//	│  • Type-safe context helpers                                   │
//	└─────────────────────────────────────────────────────────────────┘
//	                              │
//	                              ▼
//	┌─────────────────────────────────────────────────────────────────┐
//	│                    auth/adapters/stytch                         │
//	│  (Stytch-specific implementation - hidden from app layer)      │
//	└─────────────────────────────────────────────────────────────────┘
//
// # Usage
//
// In routes:
//
//	router.Use(
//	    auth.RequireAuth(authProvider),
//	    auth.RequireOrganization(orgRepo, accountRepo, logger),
//	)
//	router.GET("/resource", auth.RequirePermission("resource", "view"), handler)
//
// In handlers:
//
//	func Handler(c *gin.Context) {
//	    reqCtx := auth.GetRequestContext(c)
//	    orgID := reqCtx.OrganizationID  // int32, type-safe
//	    accountID := reqCtx.AccountID   // int32, type-safe
//	}
//
// # Adding a New Auth Provider
//
// To add a new authentication provider (e.g., Auth0, Firebase):
//
//  1. Create a new adapter in auth/adapters/<provider>/
//  2. Implement the AuthProvider interface
//  3. Map provider-specific claims to auth.Identity
//  4. Register the adapter in the DI container
//
// See auth/adapters/stytch/ for a reference implementation.
package auth

import (
	"context"
	"time"
)

// AuthProvider abstracts the authentication provider (Stytch, Auth0, Firebase, etc.).
//
// Implementations must:
//   - Verify the token signature and validity
//   - Extract user identity information
//   - Derive permissions from roles (if applicable)
//   - Return appropriate errors for invalid/expired tokens
//
// The application layer should only depend on this interface, never on
// provider-specific implementations.
type AuthProvider interface {
	// VerifyToken validates the provided token and returns the user's identity.
	//
	// The token is typically a JWT from the Authorization header.
	// Returns ErrInvalidToken, ErrTokenExpired, or other auth errors on failure.
	VerifyToken(ctx context.Context, token string) (*Identity, error)
}

// Identity represents an authenticated user in a provider-agnostic way.
//
// This struct contains all the information needed by the application
// after a user has been authenticated. Provider-specific data is stored
// in the Raw field for debugging or advanced use cases.
type Identity struct {
	// UserID is the unique identifier for the user from the auth provider.
	// For local auth, this is the database user ID. For Stytch, this is the member_id.
	UserID string `json:"user_id"`

	// Email is the user's email address.
	Email string `json:"email"`

	// EmailVerified indicates whether the email has been verified.
	EmailVerified bool `json:"email_verified"`

	// OrganizationID is the auth provider's organization/tenant identifier.
	// Optional for B2C applications. Required for B2B multi-tenant apps.
	// Use RequestContext.OrganizationID for the database ID.
	OrganizationID string `json:"organization_id,omitempty"`

	// Roles contains the user's role assignments (e.g., "admin", "user").
	Roles []Role `json:"roles"`

	// Permissions contains the derived permissions in "resource:action" format.
	// These are derived from roles by the auth provider or adapter.
	Permissions []Permission `json:"permissions"`

	// ExpiresAt is when the token/session expires.
	ExpiresAt time.Time `json:"expires_at"`

	// Raw contains provider-specific data for debugging or advanced use cases.
	// This should NOT be used in normal application logic.
	Raw map[string]any `json:"raw,omitempty"`
}

// HasRole checks if the identity has a specific role.
func (i *Identity) HasRole(role Role) bool {
	for _, r := range i.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasPermission checks if the identity has a specific permission.
func (i *Identity) HasPermission(permission Permission) bool {
	for _, p := range i.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasResourcePermission checks if the identity has permission for a resource and action.
func (i *Identity) HasResourcePermission(resource, action string) bool {
	return i.HasPermission(NewPermission(resource, action))
}

// RequestContext holds the resolved database IDs for the current request.
//
// This is set by the RequireOrganization middleware after looking up
// the organization and account in the database using the Identity's
// provider-specific IDs.
//
// Use auth.GetRequestContext(c) to retrieve this in handlers.
type RequestContext struct {
	// Identity contains the authenticated user information from the auth provider.
	Identity *Identity `json:"identity"`

	// OrganizationID is the database primary key (int32) for the organization.
	// This is resolved from Identity.OrganizationID by the middleware.
	OrganizationID int32 `json:"organization_id"`

	// AccountID is the database primary key (int32) for the user's account.
	// This is resolved from Identity.Email by the middleware.
	AccountID int32 `json:"account_id"`

	// ProviderOrgID preserves the original provider organization ID for reference.
	// Use this when making calls back to the auth provider.
	ProviderOrgID string `json:"provider_org_id,omitempty"`
}

// OrganizationRepository defines the interface for looking up organizations.
//
// This is used by the RequireOrganization middleware to resolve
// the auth provider's organization ID to a database ID.
type OrganizationRepository interface {
	// GetByProviderID looks up an organization by the auth provider's organization ID.
	// Returns the organization with its database ID, or an error if not found.
	GetByProviderID(ctx context.Context, providerOrgID string) (*Organization, error)
}

// AccountRepository defines the interface for looking up accounts.
//
// This is used by the RequireOrganization middleware to resolve
// the user's email to a database account ID within an organization.
type AccountRepository interface {
	// GetByEmail looks up an account by email within an organization.
	// Returns the account with its database ID, or an error if not found.
	GetByEmail(ctx context.Context, orgID int32, email string) (*Account, error)
}

// Organization represents the minimal organization data needed by the auth package.
type Organization struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

// Account represents the minimal account data needed by the auth package.
type Account struct {
	ID    int32  `json:"id"`
	Email string `json:"email"`
}
