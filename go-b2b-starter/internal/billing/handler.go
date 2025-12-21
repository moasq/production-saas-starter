package subscriptions

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	billingServices "github.com/moasq/go-b2b-starter/internal/billing/app/services"
	"github.com/moasq/go-b2b-starter/internal/billing/domain"
	"github.com/moasq/go-b2b-starter/internal/auth"
	"github.com/moasq/go-b2b-starter/pkg/httperr"
	"github.com/moasq/go-b2b-starter/internal/logger"
)

type Handler struct {
	billingService billingServices.BillingService
	logger         logger.Logger
}

func NewHandler(billingService billingServices.BillingService, log logger.Logger) *Handler {
	return &Handler{
		billingService: billingService,
		logger:         log,
	}
}

// GetBillingStatus godoc
// @Summary Get current billing and quota status
// @Description Retrieve the current subscription billing status and invoice quota information for the organization
// @Tags subscriptions
// @Accept json
// @Produce json
// @Success 200 {object} domain.BillingStatus "Current billing and quota status"
// @Failure 400 {object} errors.HTTPError "Invalid request parameters or missing organization context"
// @Failure 500 {object} errors.HTTPError "Internal server error"
// @Router /api/subscriptions/status [get]
func (h *Handler) GetBillingStatus(c *gin.Context) {
	reqCtx := auth.GetRequestContext(c)
	if reqCtx == nil {
		c.JSON(http.StatusBadRequest, errors.NewHTTPError(
			http.StatusBadRequest,
			"missing_context",
			"Organization context is required",
		))
		return
	}

	// Call service layer to get billing status
	billingStatus, err := h.billingService.GetBillingStatus(c.Request.Context(), reqCtx.OrganizationID)
	if err != nil {
		// Check if subscription not found - this is not necessarily an error
		// Organization might not have a subscription yet
		if err == domain.ErrSubscriptionNotFound {
			// Return a response indicating no active subscription
			c.JSON(http.StatusOK, domain.BillingStatus{
				OrganizationID:        reqCtx.OrganizationID,
				HasActiveSubscription: false,
				CanProcessInvoices:    false,
				InvoiceCount:          0,
				Reason:                "No active subscription found",
				CheckedAt:             time.Now(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, errors.NewHTTPError(
			http.StatusInternalServerError,
			"billing_status_failed",
			fmt.Sprintf("Failed to retrieve billing status: %v", err),
		))
		return
	}

	c.JSON(http.StatusOK, billingStatus)
}

// VerifyPaymentRequest represents the request payload for verifying a payment
type VerifyPaymentRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

// VerifyPayment godoc
// @Summary Verify payment from checkout session
// @Description Verifies a payment by checking the Polar checkout session and updates subscription status. This is the primary mechanism for "Verification on Redirect" pattern when user returns from payment page.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body VerifyPaymentRequest true "Checkout session ID"
// @Success 200 {object} domain.BillingStatus "Verification result with updated billing status"
// @Failure 400 {object} errors.HTTPError "Invalid request parameters or checkout session failed"
// @Failure 404 {object} errors.HTTPError "Checkout session not found"
// @Failure 500 {object} errors.HTTPError "Internal server error"
// @Router /api/subscriptions/verify-payment [post]
func (h *Handler) VerifyPayment(c *gin.Context) {
	h.logger.Info("[VerifyPayment] Starting payment verification request", nil)

	// Bind request
	var req VerifyPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("[VerifyPayment] Failed to bind request JSON", map[string]any{
			"error": err.Error(),
		})
		c.JSON(http.StatusBadRequest, errors.NewHTTPError(
			http.StatusBadRequest,
			"invalid_request",
			fmt.Sprintf("Invalid request: %v", err),
		))
		return
	}

	h.logger.Info("[VerifyPayment] Request parsed successfully", map[string]any{
		"session_id": req.SessionID,
	})

	// Validate session_id is not empty
	if req.SessionID == "" {
		h.logger.Warn("[VerifyPayment] Missing session_id in request", nil)
		c.JSON(http.StatusBadRequest, errors.NewHTTPError(
			http.StatusBadRequest,
			"missing_session_id",
			"Checkout session ID is required",
		))
		return
	}

	h.logger.Info("[VerifyPayment] Calling billing service to verify payment", map[string]any{
		"session_id": req.SessionID,
	})

	// Call service to verify payment
	billingStatus, err := h.billingService.VerifyPaymentFromCheckout(c.Request.Context(), req.SessionID)
	if err != nil {
		// Check if it's a checkout session not found error
		if err.Error() == "checkout session not found: "+req.SessionID {
			h.logger.Warn("[VerifyPayment] Checkout session not found", map[string]any{
				"session_id": req.SessionID,
			})
			c.JSON(http.StatusNotFound, errors.NewHTTPError(
				http.StatusNotFound,
				"session_not_found",
				fmt.Sprintf("Checkout session not found: %s", req.SessionID),
			))
			return
		}

		h.logger.Error("[VerifyPayment] Failed to verify payment", map[string]any{
			"session_id": req.SessionID,
			"error":      err.Error(),
		})
		c.JSON(http.StatusInternalServerError, errors.NewHTTPError(
			http.StatusInternalServerError,
			"verification_failed",
			fmt.Sprintf("Failed to verify payment: %v", err),
		))
		return
	}

	h.logger.Info("[VerifyPayment] Billing service returned status", map[string]any{
		"session_id":              req.SessionID,
		"has_active_subscription": billingStatus.HasActiveSubscription,
		"can_process_invoices":    billingStatus.CanProcessInvoices,
		"invoice_count":           billingStatus.InvoiceCount,
		"reason":                  billingStatus.Reason,
	})

	// If checkout session is not succeeded, return 400 with reason
	if !billingStatus.HasActiveSubscription && billingStatus.Reason != "Payment verified successfully" {
		h.logger.Warn("[VerifyPayment] Payment not completed", map[string]any{
			"session_id": req.SessionID,
			"reason":     billingStatus.Reason,
		})
		c.JSON(http.StatusBadRequest, errors.NewHTTPError(
			http.StatusBadRequest,
			"payment_not_completed",
			billingStatus.Reason,
		))
		return
	}

	h.logger.Info("[VerifyPayment] Payment verification completed successfully", map[string]any{
		"session_id":      req.SessionID,
		"organization_id": billingStatus.OrganizationID,
		"invoice_count":   billingStatus.InvoiceCount,
	})

	c.JSON(http.StatusOK, billingStatus)
}
