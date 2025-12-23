package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/moasq/go-b2b-starter/pkg/response"
)

// Handler handles RBAC API endpoints
type Handler struct {
	service RBACService
}

func NewHandler(service RBACService) *Handler {
	return &Handler{
		service: service,
	}
}

// GetRoles godoc
// @Summary Get all roles with permissions
// @Description Returns all available roles in the system with their associated permissions. This is the single source of truth for frontend role/permission discovery.
// @Tags RBAC
// @Produce json
// @Success 200 {object} RolesResponse "Roles with permissions"
// @Failure 500 {object} map[string]string "Internal error"
// @Router /rbac/roles [get]
func (h *Handler) GetRoles(c *gin.Context) {
	roles := h.service.GetAllRoles()

	roleDTOs := make([]RoleDTO, len(roles))
	for i, role := range roles {
		roleDTOs[i] = NewRoleDTO(role)
	}

	response.Success(c, http.StatusOK, RolesResponse{
		Roles: roleDTOs,
	})
}

// GetPermissions godoc
// @Summary Get all permissions
// @Description Returns all available permissions in the system. Each permission includes resource, action, display name, and description for frontend rendering.
// @Tags RBAC
// @Produce json
// @Success 200 {object} PermissionsResponse "All permissions"
// @Failure 500 {object} map[string]string "Internal error"
// @Router /rbac/permissions [get]
func (h *Handler) GetPermissions(c *gin.Context) {
	permissions := h.service.GetAllPermissions()

	permDTOs := make([]PermissionDTO, len(permissions))
	for i, perm := range permissions {
		permDTOs[i] = NewPermissionDTO(perm)
	}

	response.Success(c, http.StatusOK, PermissionsResponse{
		Permissions: permDTOs,
	})
}

// GetPermissionsByCategory godoc
// @Summary Get permissions grouped by category
// @Description Returns all permissions organized by their category for better UI organization.
// @Tags RBAC
// @Produce json
// @Success 200 {object} PermissionsByCategoryResponse "Permissions by category"
// @Failure 500 {object} map[string]string "Internal error"
// @Router /rbac/permissions/by-category [get]
func (h *Handler) GetPermissionsByCategory(c *gin.Context) {
	categoriesMap := h.service.GetPermissionsByCategory()

	// Convert to DTO format
	result := make(map[string][]PermissionDTO)
	for category, perms := range categoriesMap {
		permDTOs := make([]PermissionDTO, len(perms))
		for i, perm := range perms {
			permDTOs[i] = NewPermissionDTO(perm)
		}
		result[category] = permDTOs
	}

	response.Success(c, http.StatusOK, PermissionsByCategoryResponse{
		Categories: result,
	})
}

// GetRoleDetails godoc
// @Summary Get detailed information about a specific role
// @Description Returns comprehensive information about a role including permissions, statistics, and restrictions.
// @Tags RBAC
// @Produce json
// @Param role_id path string true "Role ID (member, approver, admin)"
// @Success 200 {object} RolePermissionsResponse "Role details with statistics"
// @Failure 400 {object} map[string]string "Invalid role ID"
// @Failure 404 {object} map[string]string "Role not found"
// @Router /rbac/roles/{role_id} [get]
func (h *Handler) GetRoleDetails(c *gin.Context) {
	roleID := c.Param("role_id")

	if roleID == "" {
		response.Error(c, http.StatusBadRequest, "role_id_required", nil)
		return
	}

	roleResp := NewRolePermissionsResponse(roleID)
	if roleResp == nil {
		response.Error(c, http.StatusNotFound, "role_not_found", nil)
		return
	}

	response.Success(c, http.StatusOK, roleResp)
}

// CheckPermission godoc
// @Summary Check if a role has a specific permission
// @Description Verifies whether a role has been granted a specific permission. Useful for conditional UI rendering.
// @Tags RBAC
// @Accept json
// @Produce json
// @Param body body PermissionCheckRequest true "Role and permission to check"
// @Success 200 {object} PermissionCheckResponse "Permission check result"
// @Failure 400 {object} map[string]string "Invalid request"
// @Router /rbac/check-permission [post]
func (h *Handler) CheckPermission(c *gin.Context) {
	var req PermissionCheckRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_request", err)
		return
	}

	if req.RoleID == "" || req.PermissionID == "" {
		response.Error(c, http.StatusBadRequest, "missing_parameters", nil)
		return
	}

	hasPermission := h.service.HasPermission(req.RoleID, req.PermissionID)

	response.Success(c, http.StatusOK, PermissionCheckResponse{
		RoleID:        req.RoleID,
		PermissionID:  req.PermissionID,
		HasPermission: hasPermission,
	})
}

// GetMetadata godoc
// @Summary Get RBAC system metadata
// @Description Returns summary information about the RBAC system including total roles, permissions, and categories.
// @Tags RBAC
// @Produce json
// @Success 200 {object} RBACMetadata "RBAC system metadata"
// @Router /rbac/metadata [get]
func (h *Handler) GetMetadata(c *gin.Context) {
	metadata := h.service.GetRBACMetadata()
	response.Success(c, http.StatusOK, metadata)
}
