package stytch

import (
	"context"
	"fmt"
	"time"

	"github.com/moasq/go-b2b-starter/internal/modules/auth"
	"github.com/moasq/go-b2b-starter/internal/platform/logger"
)

// MockAuthAdapter is a development-only auth adapter that bypasses Stytch.
//
// WARNING: This should NEVER be used in production. It accepts any token
// and returns a mock identity. It's only for local development when
// Stytch credentials are not configured.
type MockAuthAdapter struct {
	logger logger.Logger
}

// Ensure MockAuthAdapter implements auth.AuthProvider.
var _ auth.AuthProvider = (*MockAuthAdapter)(nil)

func NewMockAuthAdapter(log logger.Logger) *MockAuthAdapter {
	return &MockAuthAdapter{
		logger: log,
	}
}

// VerifyToken accepts any token and returns a mock identity.
// This is for development only and should never be used in production.
func (m *MockAuthAdapter) VerifyToken(ctx context.Context, token string) (*auth.Identity, error) {
	m.logger.Warn("Using mock auth adapter - accepting any token", map[string]any{
		"warning": "This is for development only. Configure real Stytch credentials for production.",
	})

	// Return a mock identity for development
	return &auth.Identity{
		UserID:         "mock-user-123",
		Email:          "dev@example.com",
		EmailVerified:  true,
		OrganizationID: "mock-org-stytch-id",
		Roles: []auth.Role{
			auth.RoleOwner,
			auth.RoleAdmin,
		},
		Permissions: []auth.Permission{
			auth.NewPermission("*", "*"), // Wildcard permission for development
		},
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Raw: map[string]any{
			"mock":       true,
			"session_id": "mock-session-123",
			"member_id":  "mock-member-123",
		},
	}, nil
}

// GetRolePermissions returns empty permissions for the mock adapter.
func (m *MockAuthAdapter) GetRolePermissions(ctx context.Context, roleID string) ([]auth.Permission, error) {
	m.logger.Debug("Mock adapter returning all permissions for role", map[string]any{
		"role_id": roleID,
	})

	// Return wildcard permission for development
	return []auth.Permission{
		auth.NewPermission("*", "*"),
	}, nil
}

// ValidatePermission always returns true in mock mode.
func (m *MockAuthAdapter) ValidatePermission(ctx context.Context, identity *auth.Identity, resource, action string) error {
	m.logger.Debug("Mock adapter allowing all permissions", map[string]any{
		"resource": resource,
		"action":   action,
		"user_id":  identity.UserID,
	})
	return nil
}

// RefreshSession is not implemented in mock mode.
func (m *MockAuthAdapter) RefreshSession(ctx context.Context, sessionToken string) (*auth.Identity, error) {
	return nil, fmt.Errorf("mock adapter: RefreshSession not implemented")
}
