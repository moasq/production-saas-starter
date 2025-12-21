package organizations

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/moasq/go-b2b-starter/internal/organizations/app/services"
	"github.com/moasq/go-b2b-starter/internal/organizations/domain"
	"github.com/moasq/go-b2b-starter/pkg/response"
	"github.com/moasq/go-b2b-starter/internal/auth"
	"github.com/moasq/go-b2b-starter/internal/logger"
)

type OrganizationHandler struct {
	orgService services.OrganizationService
	logger     logger.Logger
}

func NewOrganizationHandler(orgService services.OrganizationService, logger logger.Logger) *OrganizationHandler {
	return &OrganizationHandler{
		orgService: orgService,
		logger:     logger,
	}
}

// CreateOrganization creates a new organization
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	var req services.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request payload", map[string]interface{}{"error": err.Error()})
		response.Error(c, http.StatusBadRequest, "invalid request payload", err)
		return
	}

	org, err := h.orgService.CreateOrganization(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("failed to create organization", map[string]interface{}{"error": err.Error()})
		response.Error(c, http.StatusInternalServerError, "failed to create organization", err)
		return
	}

	response.Success(c, http.StatusCreated, org)
}

// GetOrganization gets the current organization (from context)
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	reqCtx := auth.GetRequestContext(c)
	if reqCtx == nil {
		h.logger.Error("missing request context", nil)
		response.Error(c, http.StatusBadRequest, "organization context is required", nil)
		return
	}

	org, err := h.orgService.GetOrganization(c.Request.Context(), reqCtx.OrganizationID)
	if err != nil {
		if err == domain.ErrOrganizationNotFound {
			response.Error(c, http.StatusNotFound, "organization not found", err)
			return
		}
		h.logger.Error("failed to get organization", map[string]interface{}{"org_id": reqCtx.OrganizationID, "error": err.Error()})
		response.Error(c, http.StatusInternalServerError, "failed to get organization", err)
		return
	}

	response.Success(c, http.StatusOK, org)
}

// GetOrganizationBySlug gets an organization by slug
func (h *OrganizationHandler) GetOrganizationBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		response.Error(c, http.StatusBadRequest, "slug is required", nil)
		return
	}

	org, err := h.orgService.GetOrganizationBySlug(c.Request.Context(), slug)
	if err != nil {
		if err == domain.ErrOrganizationNotFound {
			response.Error(c, http.StatusNotFound, "organization not found", err)
			return
		}
		h.logger.Error("failed to get organization by slug", map[string]interface{}{"slug": slug, "error": err.Error()})
		response.Error(c, http.StatusInternalServerError, "failed to get organization", err)
		return
	}

	response.Success(c, http.StatusOK, org)
}

// UpdateOrganization updates the current organization (from context)
func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
	reqCtx := auth.GetRequestContext(c)
	if reqCtx == nil {
		h.logger.Error("missing request context", nil)
		response.Error(c, http.StatusBadRequest, "organization context is required", nil)
		return
	}

	var req services.UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request payload", map[string]interface{}{"error": err.Error()})
		response.Error(c, http.StatusBadRequest, "invalid request payload", err)
		return
	}

	org, err := h.orgService.UpdateOrganization(c.Request.Context(), reqCtx.OrganizationID, &req)
	if err != nil {
		if err == domain.ErrOrganizationNotFound {
			response.Error(c, http.StatusNotFound, "organization not found", err)
			return
		}
		h.logger.Error("failed to update organization", map[string]interface{}{"org_id": reqCtx.OrganizationID, "error": err.Error()})
		response.Error(c, http.StatusInternalServerError, "failed to update organization", err)
		return
	}

	response.Success(c, http.StatusOK, org)
}

// ListOrganizations lists organizations with pagination
func (h *OrganizationHandler) ListOrganizations(c *gin.Context) {
	var req services.ListOrganizationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("invalid query parameters", map[string]interface{}{"error": err.Error()})
		response.Error(c, http.StatusBadRequest, "invalid query parameters", err)
		return
	}

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 10
	}

	orgResponse, err := h.orgService.ListOrganizations(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("failed to list organizations", map[string]interface{}{"error": err.Error()})
		response.Error(c, http.StatusInternalServerError, "failed to list organizations", err)
		return
	}

	response.Success(c, http.StatusOK, orgResponse)
}

// GetOrganizationStats gets statistics for the current organization (from context)
func (h *OrganizationHandler) GetOrganizationStats(c *gin.Context) {
	reqCtx := auth.GetRequestContext(c)
	if reqCtx == nil {
		h.logger.Error("missing request context", nil)
		response.Error(c, http.StatusBadRequest, "organization context is required", nil)
		return
	}

	stats, err := h.orgService.GetOrganizationStats(c.Request.Context(), reqCtx.OrganizationID)
	if err != nil {
		if err == domain.ErrOrganizationNotFound {
			response.Error(c, http.StatusNotFound, "organization not found", err)
			return
		}
		h.logger.Error("failed to get organization stats", map[string]interface{}{"org_id": reqCtx.OrganizationID, "error": err.Error()})
		response.Error(c, http.StatusInternalServerError, "failed to get organization stats", err)
		return
	}

	response.Success(c, http.StatusOK, stats)
}
