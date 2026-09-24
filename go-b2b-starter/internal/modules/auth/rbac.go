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
	PermOrgView   = NewPermission("org", "view")
	PermOrgManage = NewPermission("org", "manage")
)

// The catalog describes the starter's recommended Stytch policy. Actual request
// permissions always come from the provider; this metadata does not grant access.
var AllPermissions = []Permission{PermOrgView, PermOrgManage}

type RoleInfo struct {
	ID          string
	Name        string
	Description string
	Permissions []Permission
}

var (
	RoleMemberInfo  = RoleInfo{"member", "Member", "View the workspace.", []Permission{PermOrgView}}
	RoleManagerInfo = RoleInfo{"manager", "Manager", "View the workspace. Extend this role for your domain.", []Permission{PermOrgView}}
	RoleAdminInfo   = RoleInfo{"admin", "Admin", "Manage the workspace, members and billing.", []Permission{PermOrgView, PermOrgManage}}
)
var AllRoles = []RoleInfo{RoleMemberInfo, RoleManagerInfo, RoleAdminInfo}

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
		CanApprove:       false,
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
		Description:       "Recommended Stytch policy: organization view and manage permissions",
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
