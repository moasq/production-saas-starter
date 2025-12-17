package organizations

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/moasq/go-b2b-starter/app/organizations/app/services"
	"github.com/moasq/go-b2b-starter/app/organizations/domain"
	"github.com/moasq/go-b2b-starter/pkg/api/response"
	"github.com/moasq/go-b2b-starter/pkg/auth"
	"github.com/moasq/go-b2b-starter/pkg/logger"
)

type AccountHandler struct {
	orgService services.OrganizationService
	logger     logger.Logger
}

func NewAccountHandler(orgService services.OrganizationService, logger logger.Logger) *AccountHandler {
	return &AccountHandler{
		orgService: orgService,
		logger:     logger,
	}
}

// CreateAccount creates a new account in an organization
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	reqCtx := auth.GetRequestContext(c)
	if reqCtx == nil {
		h.logger.Error("missing request context", nil)
		response.Error(c, http.StatusBadRequest, "organization context is required", nil)
		return
	}

	var req services.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request payload", map[string]interface{}{"error": err.Error()})
		response.Error(c, http.StatusBadRequest, "invalid request payload", err)
		return
	}

	domainReq := &req
	account, err := h.orgService.CreateAccount(c.Request.Context(), reqCtx.OrganizationID, domainReq)
	if err != nil {
		if err == domain.ErrOrganizationNotFound {
			response.Error(c, http.StatusNotFound, "organization not found", err)
			return
		}
		h.logger.Error("failed to create account", map[string]interface{}{"org_id": reqCtx.OrganizationID, "error": err.Error()})
		response.Error(c, http.StatusInternalServerError, "failed to create account", err)
		return
	}

	response.Success(c, http.StatusCreated, account)
}

// GetAccount gets an account by ID
func (h *AccountHandler) GetAccount(c *gin.Context) {
	reqCtx := auth.GetRequestContext(c)
	if reqCtx == nil {
		h.logger.Error("missing request context", nil)
		response.Error(c, http.StatusBadRequest, "organization context is required", nil)
		return
	}

	// Extract account_id from path parameter
	accountIDParam := c.Param("id")
	var accountID int32
	if _, err := fmt.Sscanf(accountIDParam, "%d", &accountID); err != nil {
		h.logger.Error("invalid account ID", map[string]interface{}{"id": accountIDParam, "error": err.Error()})
		response.Error(c, http.StatusBadRequest, "invalid account ID format", err)
		return
	}

	account, err := h.orgService.GetAccount(c.Request.Context(), reqCtx.OrganizationID, accountID)
	if err != nil {
		if err == domain.ErrAccountNotFound {
			response.Error(c, http.StatusNotFound, "account not found", err)
			return
		}
		h.logger.Error("failed to get account", map[string]interface{}{"org_id": reqCtx.OrganizationID, "account_id": accountID, "error": err.Error()})
		response.Error(c, http.StatusInternalServerError, "failed to get account", err)
		return
	}

	response.Success(c, http.StatusOK, account)
}

// GetAccountByEmail gets an account by email
func (h *AccountHandler) GetAccountByEmail(c *gin.Context) {
	reqCtx := auth.GetRequestContext(c)
	if reqCtx == nil {
		h.logger.Error("missing request context", nil)
		response.Error(c, http.StatusBadRequest, "organization context is required", nil)
		return
	}

	email := c.Query("email")
	if email == "" {
		response.Error(c, http.StatusBadRequest, "email query parameter is required", nil)
		return
	}

	account, err := h.orgService.GetAccountByEmail(c.Request.Context(), reqCtx.OrganizationID, email)
	if err != nil {
		if err == domain.ErrAccountNotFound {
			response.Error(c, http.StatusNotFound, "account not found", err)
			return
		}
		h.logger.Error("failed to get account by email", map[string]interface{}{"org_id": reqCtx.OrganizationID, "email": email, "error": err.Error()})
		response.Error(c, http.StatusInternalServerError, "failed to get account", err)
		return
	}

	response.Success(c, http.StatusOK, account)
}

// ListAccounts lists all accounts in an organization
func (h *AccountHandler) ListAccounts(c *gin.Context) {
	reqCtx := auth.GetRequestContext(c)
	if reqCtx == nil {
		h.logger.Error("missing request context", nil)
		response.Error(c, http.StatusBadRequest, "organization context is required", nil)
		return
	}

	accounts, err := h.orgService.ListAccounts(c.Request.Context(), reqCtx.OrganizationID)
	if err != nil {
		if err == domain.ErrOrganizationNotFound {
			response.Error(c, http.StatusNotFound, "organization not found", err)
			return
		}
		h.logger.Error("failed to list accounts", map[string]interface{}{"org_id": reqCtx.OrganizationID, "error": err.Error()})
		response.Error(c, http.StatusInternalServerError, "failed to list accounts", err)
		return
	}

	response.Success(c, http.StatusOK, accounts)
}

// UpdateAccount updates an account
func (h *AccountHandler) UpdateAccount(c *gin.Context) {
	reqCtx := auth.GetRequestContext(c)
	if reqCtx == nil {
		h.logger.Error("missing request context", nil)
		response.Error(c, http.StatusBadRequest, "organization context is required", nil)
		return
	}

	// Extract account_id from path parameter
	accountIDParam := c.Param("id")
	var accountID int32
	if _, err := fmt.Sscanf(accountIDParam, "%d", &accountID); err != nil {
		h.logger.Error("invalid account ID", map[string]interface{}{"id": accountIDParam, "error": err.Error()})
		response.Error(c, http.StatusBadRequest, "invalid account ID format", err)
		return
	}

	var req services.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request payload", map[string]interface{}{"error": err.Error()})
		response.Error(c, http.StatusBadRequest, "invalid request payload", err)
		return
	}

	domainReq := &req
	account, err := h.orgService.UpdateAccount(c.Request.Context(), reqCtx.OrganizationID, accountID, domainReq)
	if err != nil {
		if err == domain.ErrAccountNotFound {
			response.Error(c, http.StatusNotFound, "account not found", err)
			return
		}
		h.logger.Error("failed to update account", map[string]interface{}{"org_id": reqCtx.OrganizationID, "account_id": accountID, "error": err.Error()})
		response.Error(c, http.StatusInternalServerError, "failed to update account", err)
		return
	}

	response.Success(c, http.StatusOK, account)
}

func (h *AccountHandler) DeleteAccount(c *gin.Context) {
	reqCtx := auth.GetRequestContext(c)
	if reqCtx == nil {
		h.logger.Error("missing request context", nil)
		response.Error(c, http.StatusBadRequest, "organization context is required", nil)
		return
	}

	// Extract account_id from path parameter
	accountIDParam := c.Param("id")
	var accountID int32
	if _, err := fmt.Sscanf(accountIDParam, "%d", &accountID); err != nil {
		h.logger.Error("invalid account ID", map[string]interface{}{"id": accountIDParam, "error": err.Error()})
		response.Error(c, http.StatusBadRequest, "invalid account ID format", err)
		return
	}

	err := h.orgService.DeleteAccount(c.Request.Context(), reqCtx.OrganizationID, accountID)
	if err != nil {
		if err == domain.ErrAccountNotFound {
			response.Error(c, http.StatusNotFound, "account not found", err)
			return
		}
		h.logger.Error("failed to delete account", map[string]interface{}{"org_id": reqCtx.OrganizationID, "account_id": accountID, "error": err.Error()})
		response.Error(c, http.StatusInternalServerError, "failed to delete account", err)
		return
	}

	response.Success(c, http.StatusNoContent, nil)
}

// UpdateAccountLastLogin updates account last login timestamp
func (h *AccountHandler) UpdateAccountLastLogin(c *gin.Context) {
	reqCtx := auth.GetRequestContext(c)
	if reqCtx == nil {
		h.logger.Error("missing request context", nil)
		response.Error(c, http.StatusBadRequest, "organization context is required", nil)
		return
	}

	// Extract account_id from path parameter
	accountIDParam := c.Param("id")
	var accountID int32
	if _, err := fmt.Sscanf(accountIDParam, "%d", &accountID); err != nil {
		h.logger.Error("invalid account ID", map[string]interface{}{"id": accountIDParam, "error": err.Error()})
		response.Error(c, http.StatusBadRequest, "invalid account ID format", err)
		return
	}

	account, err := h.orgService.UpdateAccountLastLogin(c.Request.Context(), reqCtx.OrganizationID, accountID)
	if err != nil {
		if err == domain.ErrAccountNotFound {
			response.Error(c, http.StatusNotFound, "account not found", err)
			return
		}
		h.logger.Error("failed to update account last login", map[string]interface{}{"org_id": reqCtx.OrganizationID, "account_id": accountID, "error": err.Error()})
		response.Error(c, http.StatusInternalServerError, "failed to update account last login", err)
		return
	}

	response.Success(c, http.StatusOK, account)
}

func (h *AccountHandler) CheckAccountPermission(c *gin.Context) {
	reqCtx := auth.GetRequestContext(c)
	if reqCtx == nil {
		h.logger.Error("missing request context", nil)
		response.Error(c, http.StatusBadRequest, "organization context is required", nil)
		return
	}

	// Extract account_id from path parameter
	accountIDParam := c.Param("id")
	var accountID int32
	if _, err := fmt.Sscanf(accountIDParam, "%d", &accountID); err != nil {
		h.logger.Error("invalid account ID", map[string]interface{}{"id": accountIDParam, "error": err.Error()})
		response.Error(c, http.StatusBadRequest, "invalid account ID format", err)
		return
	}

	permission, err := h.orgService.CheckAccountPermission(c.Request.Context(), reqCtx.OrganizationID, accountID)
	if err != nil {
		if err == domain.ErrAccountNotFound {
			response.Error(c, http.StatusNotFound, "account not found", err)
			return
		}
		h.logger.Error("failed to check account permission", map[string]interface{}{"org_id": reqCtx.OrganizationID, "account_id": accountID, "error": err.Error()})
		response.Error(c, http.StatusInternalServerError, "failed to check account permission", err)
		return
	}

	response.Success(c, http.StatusOK, permission)
}

// GetAccountStats gets account statistics
func (h *AccountHandler) GetAccountStats(c *gin.Context) {
	// Extract account_id from path parameter
	accountIDParam := c.Param("id")
	var accountID int32
	if _, err := fmt.Sscanf(accountIDParam, "%d", &accountID); err != nil {
		h.logger.Error("invalid account ID", map[string]interface{}{"id": accountIDParam, "error": err.Error()})
		response.Error(c, http.StatusBadRequest, "invalid account ID format", err)
		return
	}

	stats, err := h.orgService.GetAccountStats(c.Request.Context(), accountID)
	if err != nil {
		if err == domain.ErrAccountNotFound {
			response.Error(c, http.StatusNotFound, "account not found", err)
			return
		}
		h.logger.Error("failed to get account stats", map[string]interface{}{"account_id": accountID, "error": err.Error()})
		response.Error(c, http.StatusInternalServerError, "failed to get account stats", err)
		return
	}

	response.Success(c, http.StatusOK, stats)
}
