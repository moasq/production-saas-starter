package paywall

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/moasq/go-b2b-starter/internal/auth"
)

// MiddlewareConfig configures the subscription middleware behavior.
type MiddlewareConfig struct {
	// ErrorHandler is called when subscription check fails.
	// If nil, default JSON responses are used.
	ErrorHandler func(c *gin.Context, statusCode int, response *ErrorResponse)

	// UpgradeURL is the URL to include in error responses for upgrading subscription.
	// Example: "/billing" or "https://app.example.com/billing"
	UpgradeURL string

	// AllowTrialing determines if trialing subscriptions are allowed.
	// Default: true (trialing is allowed)
	AllowTrialing bool
}

// DefaultMiddlewareConfig returns the default middleware configuration.
func DefaultMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		ErrorHandler:  defaultErrorHandler,
		UpgradeURL:    "/billing",
		AllowTrialing: true,
	}
}

// defaultErrorHandler sends JSON error responses.
func defaultErrorHandler(c *gin.Context, statusCode int, response *ErrorResponse) {
	c.JSON(statusCode, response)
}

// Middleware provides subscription middleware functions.
//
// Use NewMiddleware to create an instance with proper dependencies.
type Middleware struct {
	provider SubscriptionStatusProvider
	config   *MiddlewareConfig
}

// Parameters:
//   - provider: The subscription status provider (implements SubscriptionStatusProvider)
//   - config: Middleware configuration (optional, uses defaults if nil)
func NewMiddleware(provider SubscriptionStatusProvider, config *MiddlewareConfig) *Middleware {
	if config == nil {
		config = DefaultMiddlewareConfig()
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultErrorHandler
	}
	return &Middleware{
		provider: provider,
		config:   config,
	}
}

// RequireActiveSubscription returns middleware that checks subscription status.
//
// This middleware:
//  1. Gets OrganizationID from auth context (requires RequireOrganization to run first)
//  2. Checks subscription status from the SubscriptionStatusProvider
//  3. Sets SubscriptionStatus in Gin context if active
//  4. Returns 402 Payment Required if subscription is not active
//
// Must be called AFTER auth.RequireOrganization middleware.
//
// Usage:
//
//	router.Use(authMiddleware.RequireAuth())
//	router.Use(authMiddleware.RequireOrganization())
//	router.Use(subscriptionMiddleware.RequireActiveSubscription())
func (m *Middleware) RequireActiveSubscription() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip OPTIONS requests (CORS preflight)
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Get organization ID from auth context
		orgID := auth.GetOrganizationID(c)
		if orgID == 0 {
			m.config.ErrorHandler(c, http.StatusInternalServerError, &ErrorResponse{
				Error:   "configuration_error",
				Message: "Organization context required - ensure RequireOrganization middleware is applied",
			})
			c.Abort()
			return
		}

		// Check subscription status from database (fast read)
		status, err := m.provider.GetSubscriptionStatus(c.Request.Context(), orgID)
		if err != nil {
			// No subscription found
			m.config.ErrorHandler(c, http.StatusPaymentRequired, &ErrorResponse{
				Error:      "subscription_required",
				Message:    "An active subscription is required to access this feature",
				UpgradeURL: m.config.UpgradeURL,
				Status:     StatusNone,
			})
			c.Abort()
			return
		}

		// Lazy Guarding: If DB says inactive BUT subscription exists (not "none"),
		// double-check with payment provider in case we missed a webhook
		if !status.IsActive && status.Status != StatusNone {
			// Attempt to refresh subscription status from provider
			freshStatus, refreshErr := m.provider.RefreshSubscriptionStatus(c.Request.Context(), orgID)

			if refreshErr == nil && freshStatus != nil && freshStatus.IsActive {
				// Webhook was missed! Provider says active, update our status
				status = freshStatus
				// Log this occurrence for monitoring
				// Console log for lazy guard activation
				fmt.Printf("🔄 LAZY GUARD ACTIVATED - Org: %d | DB said: %s | Provider says: %s | Access granted\n",
					orgID, status.Status, freshStatus.Status)
			}
			// If refresh fails or still inactive, continue with original status
		}

		// Check if subscription is active (after potential refresh)
		if !status.IsActive {
			response := m.buildErrorResponse(status)
			m.config.ErrorHandler(c, http.StatusPaymentRequired, response)
			c.Abort()
			return
		}

		// Set subscription status in context for downstream handlers
		SetSubscriptionStatus(c, status)

		c.Next()
	}
}

// buildErrorResponse creates an appropriate error response based on subscription status.
func (m *Middleware) buildErrorResponse(status *SubscriptionStatus) *ErrorResponse {
	response := &ErrorResponse{
		UpgradeURL: m.config.UpgradeURL,
		Status:     status.Status,
	}

	switch status.Status {
	case StatusPastDue:
		response.Error = "payment_failed"
		response.Message = "Your subscription payment has failed. Please update your payment method."
	case StatusCanceled:
		response.Error = "subscription_canceled"
		response.Message = "Your subscription has been canceled. Please resubscribe to continue."
	case StatusUnpaid:
		response.Error = "payment_required"
		response.Message = "Your subscription is unpaid. Please update your payment method."
	default:
		response.Error = "subscription_inactive"
		response.Message = "An active subscription is required to access this feature"
		if status.Reason != "" {
			response.Message = status.Reason
		}
	}

	return response
}

// RequireActiveSubscriptionFunc is a standalone middleware function.
//
// This is a convenience function that doesn't require a Middleware instance.
// It uses the default configuration for error handling.
//
// Usage:
//
//	router.GET("/ai/generate",
//	    subscription.RequireActiveSubscriptionFunc(provider),
//	    handler)
func RequireActiveSubscriptionFunc(provider SubscriptionStatusProvider) gin.HandlerFunc {
	m := NewMiddleware(provider, nil)
	return m.RequireActiveSubscription()
}

// OptionalSubscriptionStatus returns middleware that checks subscription status
// but doesn't block the request if the subscription is inactive.
//
// This middleware:
//  1. Gets OrganizationID from auth context
//  2. Checks subscription status from the SubscriptionStatusProvider
//  3. Sets SubscriptionStatus in Gin context (active or not)
//  4. Always continues to the next handler
//
// Use this when you want to know subscription status but allow access regardless.
// Handlers can then check subscription.GetSubscriptionStatus(c) to adjust behavior.
//
// Example use case: Show "upgrade" prompts to free users while still allowing access.
func (m *Middleware) OptionalSubscriptionStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip OPTIONS requests (CORS preflight)
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Get organization ID from auth context
		orgID := auth.GetOrganizationID(c)
		if orgID == 0 {
			// No org context, continue without subscription info
			c.Next()
			return
		}

		// Check subscription status (ignore errors, just set status if available)
		status, err := m.provider.GetSubscriptionStatus(c.Request.Context(), orgID)
		if err == nil && status != nil {
			SetSubscriptionStatus(c, status)
		}

		c.Next()
	}
}
