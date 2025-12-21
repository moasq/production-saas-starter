package auth

// Role represents a user role in the system.
//
// Roles are assigned to users and determine their base permissions.
// The application uses a three-tier RBAC system:
//
//   - Member: Basic access (view and create)
//   - Manager: Elevated access (edit, delete, approve)
//   - Admin: Full system control
//
// # Adding New Roles
//
// To add a new role:
//  1. Add the role constant below
//  2. Add role info in rbac.go (AllRoles)
//  3. Configure the role in your auth provider (e.g., Stytch dashboard)
//
// # Role Source of Truth
//
// The auth provider (e.g., Stytch) is the source of truth for roles at runtime.
// The definitions in rbac.go are used as a fallback when the auth provider
// doesn't provide explicit permissions.
type Role string

// Core RBAC roles.
//
// These must match the roles configured in the auth provider.
const (
	// RoleMember is for basic users with view and create access.
	// Can: View resources, create resources
	// Cannot: Edit, delete, approve, manage org
	RoleMember Role = "member"

	// RoleManager is for users with elevated access.
	// Can: View, create, edit, delete, approve resources
	// Cannot: Manage organization settings
	RoleManager Role = "manager"

	// RoleAdmin has full system control.
	// Can: Everything - no restrictions
	RoleAdmin Role = "admin"
)

// Legacy role aliases for backward compatibility.
//
// These map to the new role constants and will be removed in a future version.
const (
	// RoleOwner is a legacy alias for RoleAdmin.
	// Deprecated: Use RoleAdmin instead.
	RoleOwner Role = "owner"

	// RoleApprover is a legacy alias for RoleManager.
	// Deprecated: Use RoleManager instead.
	RoleApprover Role = "approver"

	// RoleReviewer is a legacy alias for RoleManager.
	// Deprecated: Use RoleManager instead.
	RoleReviewer Role = "reviewer"

	// RoleEmployee is a legacy alias for RoleMember.
	// Deprecated: Use RoleMember instead.
	RoleEmployee Role = "employee"
)

// String returns the string representation of the role.
func (r Role) String() string {
	return string(r)
}

// IsValid checks if the role is a known role.
func (r Role) IsValid() bool {
	normalized := NormalizeRole(string(r))
	roleInfo := GetRoleInfo(string(normalized))
	return roleInfo != nil
}

// NormalizeRole converts legacy role names to current ones.
//
// This handles backward compatibility for old role assignments:
//   - "owner" -> "admin"
//   - "approver" -> "manager"
//   - "reviewer" -> "manager"
//   - "employee" -> "member"
//   - "stytch_member" -> "member" (strips provider prefix)
func NormalizeRole(roleStr string) Role {
	role := Role(roleStr)

	// Handle provider-prefixed roles (e.g., "stytch_member")
	switch role {
	case "stytch_member":
		return RoleMember
	case "stytch_admin":
		return RoleAdmin
	case "stytch_manager":
		return RoleManager
	}

	// Handle legacy roles
	switch role {
	case RoleOwner:
		return RoleAdmin
	case RoleApprover, RoleReviewer:
		return RoleManager
	case RoleEmployee:
		return RoleMember
	}

	return role
}

// GetRolePermissions returns the default permissions for a role.
//
// Returns nil if the role is not recognized.
// This is used as a fallback when the auth provider doesn't provide permissions.
//
// Permissions are defined in rbac.go which is the single source of truth.
func GetRolePermissions(role Role) []Permission {
	// Normalize the role to handle legacy names
	normalized := NormalizeRole(string(role))

	// Get role info from rbac.go (single source of truth)
	roleInfo := GetRoleInfo(string(normalized))
	if roleInfo == nil {
		return nil
	}

	return roleInfo.Permissions
}

// HasRolePermission checks if a role has a specific permission.
//
// This uses the role definitions from rbac.go and is used as a fallback
// when the auth provider doesn't include explicit permissions.
func HasRolePermission(role Role, resource, action string) bool {
	perms := GetRolePermissions(role)
	target := NewPermission(resource, action)
	for _, p := range perms {
		if p == target {
			return true
		}
	}
	return false
}
