package auth

import (
	"fmt"
	"strings"
)

// Permission represents an authorization permission in "resource:action" format.
//
// Permissions follow the pattern "resource:action" where:
//   - resource: The entity being accessed (e.g., "invoice", "org", "approval")
//   - action: The operation being performed (e.g., "view", "create", "manage")
//
// # Examples
//
//	auth.NewPermission("invoice", "create")  // "invoice:create"
//	auth.NewPermission("org", "manage")      // "org:manage"
//	auth.NewPermission("approval", "approve") // "approval:approve"
//
// # Adding New Permissions
//
// To add a new permission:
//  1. Add it to the appropriate role in DefaultRolePermissions (roles.go)
//  2. Configure it in your auth provider (e.g., Stytch RBAC policy)
//  3. Use RequirePermission("resource", "action") in your routes
type Permission string

// NewPermission creates a permission from resource and action.
//
// Example:
//
//	perm := auth.NewPermission("invoice", "create")
//	// perm = "invoice:create"
func NewPermission(resource, action string) Permission {
	return Permission(fmt.Sprintf("%s:%s", resource, action))
}

// String returns the string representation of the permission.
func (p Permission) String() string {
	return string(p)
}

// Resource returns the resource part of the permission.
//
// Example:
//
//	perm := auth.NewPermission("invoice", "create")
//	perm.Resource() // "invoice"
func (p Permission) Resource() string {
	parts := strings.SplitN(string(p), ":", 2)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// Action returns the action part of the permission.
//
// Example:
//
//	perm := auth.NewPermission("invoice", "create")
//	perm.Action() // "create"
func (p Permission) Action() string {
	parts := strings.SplitN(string(p), ":", 2)
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// IsValid checks if the permission has both resource and action parts.
func (p Permission) IsValid() bool {
	parts := strings.SplitN(string(p), ":", 2)
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}

// Matches checks if this permission matches another permission.
//
// This is a simple equality check. For wildcard matching,
// use MatchesWithWildcard.
func (p Permission) Matches(other Permission) bool {
	return p == other
}

// MatchesWithWildcard checks if this permission matches another,
// supporting wildcards.
//
// Wildcards:
//   - "*:*" matches any permission
//   - "resource:*" matches any action on that resource
//   - "*:action" matches that action on any resource
//
// Example:
//
//	auth.Permission("invoice:*").MatchesWithWildcard("invoice:create") // true
//	auth.Permission("*:view").MatchesWithWildcard("invoice:view")      // true
func (p Permission) MatchesWithWildcard(other Permission) bool {
	if p == other {
		return true
	}

	myResource := p.Resource()
	myAction := p.Action()
	otherResource := other.Resource()
	otherAction := other.Action()

	// Check for wildcards
	resourceMatch := myResource == "*" || myResource == otherResource
	actionMatch := myAction == "*" || myAction == otherAction

	return resourceMatch && actionMatch
}

// PermissionSet is a helper for checking multiple permissions efficiently.
type PermissionSet map[Permission]struct{}

// NewPermissionSet creates a permission set from a slice of permissions.
func NewPermissionSet(permissions []Permission) PermissionSet {
	set := make(PermissionSet, len(permissions))
	for _, p := range permissions {
		set[p] = struct{}{}
	}
	return set
}

// NewPermissionSetFromStrings creates a permission set from string permissions.
func NewPermissionSetFromStrings(permissions []string) PermissionSet {
	set := make(PermissionSet, len(permissions))
	for _, p := range permissions {
		set[Permission(p)] = struct{}{}
	}
	return set
}

// Contains checks if the set contains a permission.
func (ps PermissionSet) Contains(permission Permission) bool {
	_, exists := ps[permission]
	return exists
}

// ContainsResourceAction checks if the set contains a resource:action permission.
func (ps PermissionSet) ContainsResourceAction(resource, action string) bool {
	return ps.Contains(NewPermission(resource, action))
}

// ContainsAny checks if the set contains any of the given permissions.
func (ps PermissionSet) ContainsAny(permissions ...Permission) bool {
	for _, p := range permissions {
		if ps.Contains(p) {
			return true
		}
	}
	return false
}

// ContainsAll checks if the set contains all of the given permissions.
func (ps PermissionSet) ContainsAll(permissions ...Permission) bool {
	for _, p := range permissions {
		if !ps.Contains(p) {
			return false
		}
	}
	return true
}

// ToSlice converts the permission set to a slice.
func (ps PermissionSet) ToSlice() []Permission {
	result := make([]Permission, 0, len(ps))
	for p := range ps {
		result = append(result, p)
	}
	return result
}

// PermissionsToStrings converts a slice of Permission to a slice of strings.
func PermissionsToStrings(permissions []Permission) []string {
	result := make([]string, len(permissions))
	for i, p := range permissions {
		result[i] = string(p)
	}
	return result
}

// StringsToPermissions converts a slice of strings to a slice of Permission.
func StringsToPermissions(permissions []string) []Permission {
	result := make([]Permission, len(permissions))
	for i, p := range permissions {
		result[i] = Permission(p)
	}
	return result
}

