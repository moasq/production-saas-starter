package billing

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/moasq/go-b2b-starter/internal/modules/auth"
	"github.com/moasq/go-b2b-starter/internal/modules/billing/app/services"
	"github.com/moasq/go-b2b-starter/internal/modules/billing/domain"
	platform "github.com/moasq/go-b2b-starter/internal/platform/polar"
	"github.com/moasq/go-b2b-starter/pkg/httperr"
	"net/http"
)

type Handler struct{ service services.BillingService }

func NewHandler(service services.BillingService) *Handler { return &Handler{service: service} }
func (h *Handler) GetBillingStatus(c *gin.Context) {
	ctx := auth.GetRequestContext(c)
	if ctx == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	status, err := h.service.GetBillingStatus(c.Request.Context(), ctx.OrganizationID)
	h.respond(c, status, err)
}
func (h *Handler) VerifyPayment(c *gin.Context) {
	ctx := auth.GetRequestContext(c)
	if ctx == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	var body struct {
		SessionID string `json:"session_id" binding:"required,uuid"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, httperr.NewHTTPError(http.StatusBadRequest, "invalid_request", "A valid checkout session_id is required"))
		return
	}
	status, err := h.service.VerifyPaymentFromCheckout(c.Request.Context(), ctx.OrganizationID, body.SessionID)
	h.respond(c, status, err)
}
func (h *Handler) respond(c *gin.Context, result *domain.BillingStatus, err error) {
	if err == nil {
		c.JSON(http.StatusOK, result)
		return
	}
	status, code, message := http.StatusBadGateway, "billing_unavailable", "Billing provider is temporarily unavailable"
	switch {
	case errors.Is(err, domain.ErrCheckoutOwnership):
		status, code, message = http.StatusForbidden, "checkout_forbidden", "Checkout does not belong to this organization"
	case errors.Is(err, domain.ErrPaymentPending):
		status, code, message = http.StatusConflict, "payment_pending", "Payment is still processing. Refresh in a moment."
	case errors.Is(err, platform.ErrDisabled):
		status, code, message = http.StatusServiceUnavailable, "billing_disabled", "Billing is disabled"
	case errors.Is(err, platform.ErrNotFound):
		status, code, message = http.StatusNotFound, "checkout_not_found", "Checkout was not found"
	}
	c.JSON(status, httperr.NewHTTPError(status, code, message))
}
