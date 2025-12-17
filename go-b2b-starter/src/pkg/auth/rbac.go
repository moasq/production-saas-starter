package auth

// =============================================================================
// RBAC DEFINITIONS - Roles and Permissions
// =============================================================================
//
// This file is the SINGLE SOURCE OF TRUTH for all role and permission definitions.
// Customize these for your business domain.
//
// =============================================================================

// =============================================================================
// PERMISSIONS - Customize for your business domain
// =============================================================================
//
// The default uses "resource" as a generic placeholder.
// Change "resource" to match your domain entity:
//
//   E-commerce:     "product:view", "product:create", "order:manage"
//   Healthcare:     "patient:view", "records:manage", "prescription:create"
//   Project Mgmt:   "project:view", "task:create", "task:assign"
//   CRM:            "contact:view", "deal:create", "deal:close"
//   Invoice System: "invoice:view", "invoice:create", "invoice:approve"
//
// Simply rename "resource" to your domain entity!
//
// =============================================================================

var (
	// Resource permissions - rename "resource" to your domain entity
	PermResourceView    = NewPermission("resource", "view")
	PermResourceCreate  = NewPermission("resource", "create")
	PermResourceEdit    = NewPermission("resource", "edit")
	PermResourceDelete  = NewPermission("resource", "delete")
	PermResourceApprove = NewPermission("resource", "approve")

	// Organization permissions
	PermOrgView   = NewPermission("org", "view")
	PermOrgManage = NewPermission("org", "manage")
)

// AllPermissions is the complete list of all permissions in the system.
// Update this when you add or remove permissions.
var AllPermissions = []Permission{
	PermResourceView,
	PermResourceCreate,
	PermResourceEdit,
	PermResourceDelete,
	PermResourceApprove,
	PermOrgView,
	PermOrgManage,
}

// =============================================================================
// ROLES - Customize role names and permissions as needed
// =============================================================================
//
// Default roles follow a simple hierarchy:
//   - Member: Basic access (view, create)
//   - Manager: Elevated access (edit, delete, approve)
//   - Admin: Full control (everything + org management)
//
// To customize:
//   1. Change role IDs if needed (must match Stytch configuration)
//   2. Adjust permissions for each role
//   3. Add new roles if needed
//
// =============================================================================

// RoleInfo contains complete information about a role including its permissions.
// Used for API responses and role lookups.
type RoleInfo struct {
	// ID is the unique identifier for the role (e.g., "member", "manager", "admin")
	ID string
	// Name is the display name for the role
	Name string
	// Description explains the purpose and scope of the role
	Description string
	// Permissions is the list of permissions granted to this role
	Permissions []Permission
}

var (
	// RoleMemberInfo - Basic user access
	// Typical users: Employees, staff, basic users
	RoleMemberInfo = RoleInfo{
		ID:          "member",
		Name:        "Member",
		Description: "Basic access. Can view and create resources.",
		Permissions: []Permission{
			PermResourceView,
			PermResourceCreate,
		},
	}

	// RoleManagerInfo - Elevated access with approval rights
	// Typical users: Team leads, supervisors, managers
	RoleManagerInfo = RoleInfo{
		ID:          "manager",
		Name:        "Manager",
		Description: "Elevated access. Can edit, delete, and approve resources.",
		Permissions: []Permission{
			PermResourceView,
			PermResourceCreate,
			PermResourceEdit,
			PermResourceDelete,
			PermResourceApprove,
			PermOrgView,
		},
	}

	// RoleAdminInfo - Full system control
	// Typical users: Business owners, administrators
	RoleAdminInfo = RoleInfo{
		ID:          "admin",
		Name:        "Admin",
		Description: "Full control. Can manage organization settings and users.",
		Permissions: []Permission{
			PermResourceView,
			PermResourceCreate,
			PermResourceEdit,
			PermResourceDelete,
			PermResourceApprove,
			PermOrgView,
			PermOrgManage,
		},
	}
)

// AllRoles is the complete list of all roles in the RBAC system.
// Update this when you add or remove roles.
var AllRoles = []RoleInfo{
	RoleMemberInfo,
	RoleManagerInfo,
	RoleAdminInfo,
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// GetRoleInfo retrieves role information by role ID.
// Returns nil if the role is not found.
func GetRoleInfo(roleID string) *RoleInfo {
	for i := range AllRoles {
		if AllRoles[i].ID == roleID {
			return &AllRoles[i]
		}
	}
	return nil
}

// GetRolePermissionIDs returns just the permission IDs (strings) for a given role.
// Useful for Stytch integration and API responses.
func GetRolePermissionIDs(roleID string) []string {
	role := GetRoleInfo(roleID)
	if role == nil {
		return []string{}
	}

	ids := make([]string, len(role.Permissions))
	for i, perm := range role.Permissions {
		ids[i] = string(perm)
	}
	return ids
}

// HasPermission checks if a role has a specific permission.
func HasPermission(roleID string, permission Permission) bool {
	role := GetRoleInfo(roleID)
	if role == nil {
		return false
	}

	for _, perm := range role.Permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// =============================================================================
// PERMISSION MATRIX (for reference)
// =============================================================================
//
// | Permission        | Member | Manager | Admin |
// |-------------------|--------|---------|-------|
// | resource:view     |   ✓    |    ✓    |   ✓   |
// | resource:create   |   ✓    |    ✓    |   ✓   |
// | resource:edit     |        |    ✓    |   ✓   |
// | resource:delete   |        |    ✓    |   ✓   |
// | resource:approve  |        |    ✓    |   ✓   |
// | org:view          |        |    ✓    |   ✓   |
// | org:manage        |        |         |   ✓   |
//
// Role totals:
//   - Member: 2 permissions
//   - Manager: 6 permissions
//   - Admin: 7 permissions (all)
//
// =============================================================================

// =============================================================================
// API RESPONSE TYPES (DTOs)
// =============================================================================

// PermissionDTO represents a permission in API responses
type PermissionDTO struct {
	ID          string `json:"id"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// NewPermissionDTO converts a Permission to a DTO
func NewPermissionDTO(perm Permission) PermissionDTO {
	return PermissionDTO{
		ID:       string(perm),
		Resource: perm.Resource(),
		Action:   perm.Action(),
		// Generic display name and description for simple permissions
		DisplayName: perm.Resource() + " " + perm.Action(),
		Description: "Can " + perm.Action() + " " + perm.Resource(),
		Category:    "General",
	}
}

// RoleDTO represents a role with its permissions in API responses
type RoleDTO struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Permissions []PermissionDTO `json:"permissions"`
}

// NewRoleDTO converts a RoleInfo to a DTO
func NewRoleDTO(role RoleInfo) RoleDTO {
	permDTOs := make([]PermissionDTO, len(role.Permissions))
	for i, perm := range role.Permissions {
		permDTOs[i] = NewPermissionDTO(perm)
	}

	return RoleDTO{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Permissions: permDTOs,
	}
}

// RolesResponse is the response body for GET /rbac/roles
type RolesResponse struct {
	Roles []RoleDTO `json:"roles"`
}

// PermissionsResponse is the response body for GET /rbac/permissions
type PermissionsResponse struct {
	Permissions []PermissionDTO `json:"permissions"`
}

// PermissionsByCategoryResponse is the response body for GET /rbac/permissions/by-category
type PermissionsByCategoryResponse struct {
	Categories map[string][]PermissionDTO `json:"categories"`
}

// RolePermissionsResponse contains role information with detailed metadata
type RolePermissionsResponse struct {
	Role         RoleDTO          `json:"role"`
	Statistics   RoleStatistics   `json:"statistics"`
	Restrictions RoleRestrictions `json:"restrictions"`
}

// RoleStatistics provides summary information about a role
type RoleStatistics struct {
	TotalPermissions int    `json:"total_permissions"`
	CanApprove       bool   `json:"can_approve"`
	CanManageOrg     bool   `json:"can_manage_org"`
	Description      string `json:"description"`
}

// RoleRestrictions documents what a role cannot do
type RoleRestrictions struct {
	CannotDo        []string `json:"cannot_do"`
	DataAccessLevel string   `json:"data_access_level"`
	Scope           string   `json:"scope"`
}

// NewRolePermissionsResponse creates a detailed response for a role
func NewRolePermissionsResponse(roleID string) *RolePermissionsResponse {
	role := GetRoleInfo(roleID)
	if role == nil {
		return nil
	}

	stats := RoleStatistics{
		TotalPermissions: len(role.Permissions),
		CanApprove:       HasPermission(roleID, PermResourceApprove),
		CanManageOrg:     HasPermission(roleID, PermOrgManage),
		Description:      role.Description,
	}

	// Define restrictions based on role
	var restrictions RoleRestrictions
	switch roleID {
	case "member":
		restrictions = RoleRestrictions{
			CannotDo:        []string{"Edit resources", "Delete resources", "Approve requests", "Manage organization"},
			DataAccessLevel: "Basic - view and create only",
			Scope:           "Limited to own resources",
		}
	case "manager":
		restrictions = RoleRestrictions{
			CannotDo:        []string{"Manage organization settings"},
			DataAccessLevel: "Elevated - full resource access",
			Scope:           "Team-wide access",
		}
	case "admin":
		restrictions = RoleRestrictions{
			CannotDo:        []string{},
			DataAccessLevel: "Full - all data access",
			Scope:           "Organization-wide",
		}
	default:
		restrictions = RoleRestrictions{
			CannotDo:        []string{},
			DataAccessLevel: "Unknown",
			Scope:           "Unknown",
		}
	}

	return &RolePermissionsResponse{
		Role:         NewRoleDTO(*role),
		Statistics:   stats,
		Restrictions: restrictions,
	}
}

// PermissionCheckRequest is used to verify if a role has a permission
type PermissionCheckRequest struct {
	RoleID       string `json:"role_id" binding:"required"`
	PermissionID string `json:"permission_id" binding:"required"`
}

// PermissionCheckResponse indicates whether a role has a permission
type PermissionCheckResponse struct {
	RoleID        string `json:"role_id"`
	PermissionID  string `json:"permission_id"`
	HasPermission bool   `json:"has_permission"`
}

// RBACMetadata provides summary information about the RBAC system
type RBACMetadata struct {
	TotalRoles        int            `json:"total_roles"`
	TotalPermissions  int            `json:"total_permissions"`
	PermissionsByRole map[string]int `json:"permissions_by_role"`
	Description       string         `json:"description"`
}

// NewRBACMetadata creates metadata about the RBAC system
func NewRBACMetadata() RBACMetadata {
	permsByRole := make(map[string]int)
	for _, role := range AllRoles {
		permsByRole[role.ID] = len(role.Permissions)
	}

	return RBACMetadata{
		TotalRoles:        len(AllRoles),
		TotalPermissions:  len(AllPermissions),
		PermissionsByRole: permsByRole,
		Description:       "Simple RBAC system with 3 roles (Member, Manager, Admin) and 7 generic permissions",
	}
}

// =============================================================================
// SERVICE INTERFACE AND IMPLEMENTATION
// =============================================================================

// RBACService provides business logic for RBAC operations
type RBACService interface {
	GetAllRoles() []RoleInfo
	GetRoleInfo(roleID string) *RoleInfo
	GetAllPermissions() []Permission
	GetRolePermissions(roleID string) []Permission
	GetPermissionsByCategory() map[string][]Permission
	GetPermissionsByRoleID(roleID string) []string
	HasPermission(roleID string, permissionID string) bool
	GetRBACMetadata() RBACMetadata
}

// defaultRBACService implements the RBACService interface
type defaultRBACService struct{}

func NewRBACService() RBACService {
	return &defaultRBACService{}
}

func (s *defaultRBACService) GetAllRoles() []RoleInfo {
	return AllRoles
}

func (s *defaultRBACService) GetRoleInfo(roleID string) *RoleInfo {
	return GetRoleInfo(roleID)
}

func (s *defaultRBACService) GetAllPermissions() []Permission {
	return AllPermissions
}

func (s *defaultRBACService) GetRolePermissions(roleID string) []Permission {
	role := GetRoleInfo(roleID)
	if role == nil {
		return []Permission{}
	}
	return role.Permissions
}

func (s *defaultRBACService) GetPermissionsByCategory() map[string][]Permission {
	// For simplicity, return all permissions in one "General" category
	return map[string][]Permission{
		"General": AllPermissions,
	}
}

func (s *defaultRBACService) GetPermissionsByRoleID(roleID string) []string {
	return GetRolePermissionIDs(roleID)
}

func (s *defaultRBACService) HasPermission(roleID string, permissionID string) bool {
	return HasPermission(roleID, Permission(permissionID))
}

func (s *defaultRBACService) GetRBACMetadata() RBACMetadata {
	return NewRBACMetadata()
}
