package organizations

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/moasq/go-b2b-starter/app/organizations/app/services"
	"github.com/moasq/go-b2b-starter/pkg/api/response"
	"github.com/moasq/go-b2b-starter/pkg/auth"
	"github.com/moasq/go-b2b-starter/pkg/logger"
)

type MemberHandler struct {
	memberService services.MemberService
	logger        logger.Logger
}

func NewMemberHandler(
	memberService services.MemberService,
	logger logger.Logger,
) *MemberHandler {
	return &MemberHandler{
		memberService: memberService,
		logger:        logger,
	}
}

// BootstrapOrganization creates a new organization with an admin member.
// @Summary Bootstrap organization
// @Description Creates a new organization in Stytch with an initial admin member. The admin receives a magic link invite email to complete passwordless onboarding. Organization slug is auto-generated from the organization name.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body github_com_moasq_go-b2b-starter_app_organizations_app_services.BootstrapOrganizationRequest true "Organization bootstrap request (passwordless - no password required)"
// @Success 201 {object} github_com_moasq_go-b2b-starter_app_organizations_app_services.BootstrapOrganizationResponse
// @Failure 400 {object} map[string]any "Invalid request payload"
// @Failure 500 {object} map[string]any "Failed to bootstrap organization"
// @Router /auth/signup [post]
func (h *MemberHandler) BootstrapOrganization(c *gin.Context) {
	var req services.BootstrapOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid bootstrap request payload", map[string]any{
			"error": err.Error(),
		})
		response.Error(c, http.StatusBadRequest, "invalid request payload", err)
		return
	}

	// Infrastructure layer handles slug generation and duplicate handling
	result, err := h.memberService.BootstrapOrganizationWithOwner(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("failed to bootstrap organization", map[string]any{
			"org_name": req.OrgDisplayName,
			"error":    err.Error(),
		})
		response.Error(c, http.StatusInternalServerError, "failed to bootstrap organization", err)
		return
	}

	h.logger.Info("organization bootstrapped successfully", map[string]any{
		"stytch_org_id": result.OrganizationID,
		"admin_member":  result.OwnerMemberID,
		"magic_link":    result.MagicLinkSent,
	})

	response.Success(c, http.StatusCreated, result)
}

// AddMember adds a new member to an existing organization.
// @Summary Add member to organization
// @Description Adds a new member to an existing organization with a specified role. Organization ID is automatically extracted from JWT token. Member receives a magic link invite email for passwordless authentication. Request body: {"email": "user@example.com", "name": "Full Name", "role_slug": "member"}
// @Tags auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer JWT token"
// @Param email body string true "Member email address"
// @Param name body string true "Member full name"
// @Param role_slug body string false "Role slug (defaults to 'member')"
// @Success 201 {object} github_com_moasq_go-b2b-starter_app_organizations_app_services.AddMemberResponse
// @Failure 400 {object} map[string]any "Invalid request payload or missing organization context"
// @Failure 500 {object} map[string]any "Failed to add member"
// @Router /auth/members [post]
func (h *MemberHandler) AddMember(c *gin.Context) {
	reqCtx := auth.GetRequestContext(c)
	if reqCtx == nil {
		h.logger.Error("request context not found", nil)
		response.Error(c, http.StatusBadRequest, "organization context is required", nil)
		return
	}

	var req services.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid add member request payload", map[string]any{
			"error": err.Error(),
		})
		response.Error(c, http.StatusBadRequest, "invalid request payload", err)
		return
	}

	req.OrgID = reqCtx.ProviderOrgID
	if strings.TrimSpace(req.RoleSlug) == "" {
		req.RoleSlug = "member"
	}

	result, err := h.memberService.AddMemberDirect(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("failed to add member", map[string]any{
			"org_id": reqCtx.ProviderOrgID,
			"email":  req.Email,
			"error":  err.Error(),
		})
		response.Error(c, http.StatusInternalServerError, "failed to add member", err)
		return
	}

	h.logger.Info("member added to organization", map[string]any{
		"org_id":      result.OrgID,
		"member_id":   result.MemberID,
		"invite_sent": result.InviteSent,
	})

	response.Success(c, http.StatusCreated, result)
}

// ListMembers retrieves all members of the current organization.
// @Summary List organization members
// @Description Retrieves all members of the current organization. Restricted to admin role only.
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} github_com_moasq_go-b2b-starter_app_organizations_app_services.ListMembersResponse
// @Failure 400 {object} map[string]any "Missing organization context"
// @Failure 403 {object} map[string]any "Insufficient permissions - admin role required"
// @Failure 500 {object} map[string]any "Failed to list members"
// @Router /auth/members [get]
func (h *MemberHandler) ListMembers(c *gin.Context) {
	reqCtx := auth.GetRequestContext(c)
	if reqCtx == nil {
		h.logger.Error("request context not found", nil)
		response.Error(c, http.StatusBadRequest, "organization context is required", nil)
		return
	}

	result, err := h.memberService.ListOrganizationMembers(c.Request.Context(), reqCtx.ProviderOrgID)
	if err != nil {
		h.logger.Error("failed to list members", map[string]any{
			"org_id": reqCtx.ProviderOrgID,
			"error":  err.Error(),
		})
		response.Error(c, http.StatusInternalServerError, "failed to list members", err)
		return
	}

	h.logger.Info("members listed successfully", map[string]any{
		"org_id": reqCtx.ProviderOrgID,
		"count":  result.Total,
	})

	response.Success(c, http.StatusOK, result)
}

// GetProfile retrieves the current authenticated user's profile.
// @Summary Get current user profile
// @Description Retrieves comprehensive profile information for the currently authenticated user, including member details, organization info, and account status.
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} github_com_moasq_go-b2b-starter_app_organizations_app_services.ProfileResponse
// @Failure 400 {object} map[string]any "Missing required context (organization or claims)"
// @Failure 401 {object} map[string]any "Authentication required"
// @Failure 500 {object} map[string]any "Failed to retrieve profile"
// @Router /auth/profile/me [get]
func (h *MemberHandler) GetProfile(c *gin.Context) {
	reqCtx := auth.GetRequestContext(c)
	if reqCtx == nil {
		h.logger.Error("request context not found", nil)
		response.Error(c, http.StatusBadRequest, "organization context is required", nil)
		return
	}

	identity := reqCtx.Identity
	if identity == nil {
		h.logger.Error("identity not found in context", nil)
		response.Error(c, http.StatusUnauthorized, "authentication required", nil)
		return
	}

	// Get profile using service
	profile, err := h.memberService.GetCurrentUserProfile(
		c.Request.Context(),
		reqCtx.ProviderOrgID,
		identity.UserID, // member_id
		identity.Email,
	)
	if err != nil {
		h.logger.Error("failed to get user profile", map[string]any{
			"org_id":    reqCtx.ProviderOrgID,
			"member_id": identity.UserID,
			"email":     identity.Email,
			"error":     err.Error(),
		})
		response.Error(c, http.StatusInternalServerError, "failed to retrieve profile", err)
		return
	}

	// Add computed permissions from identity (derived from Stytch RBAC policy)
	profile.Permissions = auth.PermissionsToStrings(identity.Permissions)

	h.logger.Info("profile retrieved successfully", map[string]any{
		"member_id":         identity.UserID,
		"org_id":            reqCtx.ProviderOrgID,
		"email":             identity.Email,
		"permissions_count": len(profile.Permissions),
	})

	response.Success(c, http.StatusOK, profile)
}

// @Summary Delete organization member
// @Description Removes a member from the organization (deletes from both Stytch and internal database). Only admins can delete members.
// @Tags auth
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer JWT token"
// @Param member_id path string true "Member ID to delete"
// @Success 204 {object} map[string]any "Member deleted successfully"
// @Failure 400 {object} map[string]any "Invalid member ID or missing organization context"
// @Failure 403 {object} map[string]any "Insufficient permissions - admin role required"
// @Failure 404 {object} map[string]any "Member not found"
// @Failure 500 {object} map[string]any "Failed to delete member"
// @Router /auth/members/{member_id} [delete]
func (h *MemberHandler) DeleteMember(c *gin.Context) {
	reqCtx := auth.GetRequestContext(c)
	if reqCtx == nil {
		h.logger.Error("request context not found", nil)
		response.Error(c, http.StatusBadRequest, "organization context is required", nil)
		return
	}

	identity := reqCtx.Identity
	if identity == nil {
		h.logger.Error("identity not found in context", nil)
		response.Error(c, http.StatusUnauthorized, "authentication required", nil)
		return
	}

	// Extract member_id from path parameter
	memberID := c.Param("member_id")
	if memberID == "" {
		h.logger.Error("member_id path parameter is missing", nil)
		response.Error(c, http.StatusBadRequest, "member_id is required", nil)
		return
	}

	// Business rule: Cannot delete yourself
	if memberID == identity.UserID {
		h.logger.Warn("user attempted to delete themselves", map[string]any{
			"member_id":    memberID,
			"current_user": identity.UserID,
			"org_id":       reqCtx.ProviderOrgID,
		})
		response.Error(c, http.StatusForbidden, "cannot delete yourself", nil)
		return
	}

	// Delete member using service
	err := h.memberService.DeleteOrganizationMember(c.Request.Context(), reqCtx.ProviderOrgID, memberID)
	if err != nil {
		h.logger.Error("failed to delete member", map[string]any{
			"org_id":    reqCtx.ProviderOrgID,
			"member_id": memberID,
			"error":     err.Error(),
		})
		response.Error(c, http.StatusInternalServerError, "failed to delete member", err)
		return
	}

	h.logger.Info("member deleted successfully", map[string]any{
		"member_id":  memberID,
		"org_id":     reqCtx.ProviderOrgID,
		"deleted_by": identity.UserID,
	})

	response.Success(c, http.StatusNoContent, nil)
}

// @Summary Check if email exists
// @Description Checks if an email exists in any organization. Returns 200 OK (empty response) if exists, 404 Not Found if doesn't exist. This is a public endpoint used during login flow.
// @Tags auth
// @Accept json
// @Produce json
// @Param email query string true "Email address to check"
// @Success 200 "Email exists"
// @Failure 400 {object} map[string]any "Invalid email format"
// @Failure 404 {object} map[string]any "Email not found"
// @Failure 500 {object} map[string]any "Internal server error"
// @Router /auth/check-email [get]
func (h *MemberHandler) CheckEmail(c *gin.Context) {
	// Extract and validate email from query parameter
	email := strings.TrimSpace(c.Query("email"))
	if email == "" {
		h.logger.Warn("email parameter is missing", nil)
		response.Error(c, http.StatusBadRequest, "email parameter is required", nil)
		return
	}

	h.logger.Debug("checking email existence", map[string]any{
		"email": email,
	})

	// Check if email exists using service
	exists, err := h.memberService.CheckEmailExists(c.Request.Context(), email)
	if err != nil {
		h.logger.Error("failed to check email existence", map[string]any{
			"email": email,
			"error": err.Error(),
		})
		response.Error(c, http.StatusInternalServerError, "failed to check email existence", err)
		return
	}

	// Return 404 if email doesn't exist
	if !exists {
		h.logger.Debug("email not found", map[string]any{
			"email": email,
		})
		response.Error(c, http.StatusNotFound, "email not found", nil)
		return
	}

	// Return 200 OK with empty response if email exists
	h.logger.Debug("email exists", map[string]any{
		"email": email,
	})
	response.Success(c, http.StatusOK, gin.H{})
}
