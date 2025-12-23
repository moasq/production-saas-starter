package paywall

import "errors"

// Subscription errors.
//
// These errors are returned by the subscription package and can be checked
// by application code to handle specific error cases.
var (
	// ErrNoSubscription is returned when no subscription exists for the organization.
	// HTTP status: 402 Payment Required
	ErrNoSubscription = errors.New("no subscription found")

	// ErrSubscriptionInactive is returned when the subscription exists but is not active.
	// This includes statuses like "past_due", "canceled", "unpaid".
	// HTTP status: 402 Payment Required
	ErrSubscriptionInactive = errors.New("subscription is not active")

	// ErrSubscriptionExpired is returned when the subscription's billing period has ended.
	// HTTP status: 402 Payment Required
	ErrSubscriptionExpired = errors.New("subscription has expired")

	// ErrSubscriptionCanceled is returned when the subscription has been explicitly canceled.
	// HTTP status: 402 Payment Required
	ErrSubscriptionCanceled = errors.New("subscription has been canceled")

	// ErrPaymentFailed is returned when the subscription payment has failed.
	// HTTP status: 402 Payment Required
	ErrPaymentFailed = errors.New("subscription payment failed")

	// ErrMissingOrganization is returned when organization ID is not in context.
	// This means RequireOrganization middleware hasn't run.
	// HTTP status: 500 Internal Server Error (misconfigured middleware)
	ErrMissingOrganization = errors.New("organization context required")
)

// IsPaymentRequiredError returns true if the error requires payment (402).
func IsPaymentRequiredError(err error) bool {
	return errors.Is(err, ErrNoSubscription) ||
		errors.Is(err, ErrSubscriptionInactive) ||
		errors.Is(err, ErrSubscriptionExpired) ||
		errors.Is(err, ErrSubscriptionCanceled) ||
		errors.Is(err, ErrPaymentFailed)
}

// HTTPStatusCode returns the appropriate HTTP status code for a subscription error.
//
// Returns:
//   - 402 for payment-related errors
//   - 500 for configuration errors
func HTTPStatusCode(err error) int {
	if IsPaymentRequiredError(err) {
		return 402
	}
	if errors.Is(err, ErrMissingOrganization) {
		return 500
	}
	return 500
}

// ErrorResponse represents the JSON error response for subscription errors.
//
// This is used by the middleware to return a consistent error format
// that includes helpful information for the client.
type ErrorResponse struct {
	// Error is the error code (e.g., "subscription_required", "payment_failed")
	Error string `json:"error"`

	// Message is a human-readable description of the error.
	Message string `json:"message"`

	// UpgradeURL is the URL where the user can update their subscription.
	// Optional - only included when configured.
	UpgradeURL string `json:"upgrade_url,omitempty"`

	// Status is the subscription status that caused the error.
	// Optional - helps the client understand the specific issue.
	Status string `json:"status,omitempty"`
}
