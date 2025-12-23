// Package paywall provides access gating middleware for B2B SaaS applications.
//
// This package abstracts away the billing provider (Polar, Stripe, Paddle, etc.)
// and provides a clean interface for checking subscription status before allowing
// access to protected resources. It acts as a "payment bouncer" - checking if an
// organization has an active subscription before granting access to premium features.
//
// # Architecture
//
// The paywall package follows the adapter pattern, similar to the auth package:
//
//	┌─────────────────────────────────────────────────────────────────┐
//	│                        Application Layer                        │
//	│  (handlers, services - use paywall.GetSubscriptionStatus)       │
//	└─────────────────────────────────────────────────────────────────┘
//	                              │
//	                              ▼
//	┌─────────────────────────────────────────────────────────────────┐
//	│                       paywall package                           │
//	│  • SubscriptionStatusProvider interface                         │
//	│  • SubscriptionStatus (provider-agnostic status)               │
//	│  • Middleware (RequireActiveSubscription)                       │
//	│  • Type-safe context helpers                                    │
//	└─────────────────────────────────────────────────────────────────┘
//	                              │
//	                              ▼
//	┌─────────────────────────────────────────────────────────────────┐
//	│                  app/billing module adapter                     │
//	│  (Polar/Stripe-specific - hidden from app layer)               │
//	└─────────────────────────────────────────────────────────────────┘
//
// # Event-Driven Integration
//
// The paywall package does NOT manage subscriptions directly. It only reads
// subscription status from local database. The billing module (app/billing)
// handles subscription lifecycle via webhooks and events:
//
//   - Polar/Stripe sends webhook → billing module processes it
//   - billing module updates local DB → paywall reads from DB
//   - No direct coupling between paywall and billing providers
//
// # Usage
//
// In routes:
//
//	paywallMiddleware := paywall.NewMiddleware(provider, nil)
//	router.Use(
//	    auth.RequireAuth(authProvider),
//	    auth.RequireOrganization(orgRepo, accountRepo),
//	    paywallMiddleware.RequireActiveSubscription(),
//	)
//
// In handlers:
//
//	func Handler(c *gin.Context) {
//	    status := paywall.GetSubscriptionStatus(c)
//	    if status != nil && status.IsActive {
//	        // Subscription is active
//	    }
//	}
//
// # The "Swiss Cheese" Strategy
//
// NOT all routes should require active subscription. Users with failed payments
// need access to billing/settings to fix their payment method:
//
//   - Protected routes: AI features, OCR, reports, expensive operations
//   - Unprotected routes: Billing portal, settings, profile, webhooks
package paywall

import (
	"context"
	"time"
)

// SubscriptionStatusProvider abstracts how subscription status is retrieved.
//
// The subscriptions module implements this interface. The middleware package
// doesn't know about Polar, Stripe, or any specific provider.
//
// Implementations should:
//   - Read from local database only (for speed)
//   - Never call external APIs during request handling
//   - Return appropriate errors when subscription is missing
type SubscriptionStatusProvider interface {
	// GetSubscriptionStatus checks if organization has an active subscription.
	// Returns status from local database only (fast, no external API calls).
	// The organizationID is the database primary key (int32).
	GetSubscriptionStatus(ctx context.Context, organizationID int32) (*SubscriptionStatus, error)

	// RefreshSubscriptionStatus forces a sync from the payment provider API.
	// This is the lazy guarding mechanism - used when DB says expired but we want
	// to double-check with the provider in case we missed a webhook.
	// Returns updated status after syncing with provider.
	RefreshSubscriptionStatus(ctx context.Context, organizationID int32) (*SubscriptionStatus, error)
}

// SubscriptionStatus represents the organization's billing state.
//
// This is a provider-agnostic representation of subscription status.
// The status is typically synced from the payment provider via webhooks.
type SubscriptionStatus struct {
	// OrganizationID is the database primary key for the organization.
	OrganizationID int32 `json:"organization_id"`

	// IsActive indicates whether the subscription allows access to protected features.
	// True for "active" and "trialing" statuses.
	IsActive bool `json:"is_active"`

	// Status is the raw subscription status from the provider.
	// Common values: "active", "trialing", "past_due", "canceled", "unpaid"
	Status string `json:"status"`

	// ExpiresAt is when the current billing period ends.
	// After this time, the subscription may need renewal.
	ExpiresAt time.Time `json:"expires_at,omitempty"`

	// Reason provides a human-readable explanation when IsActive is false.
	// Examples: "subscription expired", "payment failed", "no subscription found"
	Reason string `json:"reason,omitempty"`
}

// IsTrialing returns true if the subscription is in a trial period.
func (s *SubscriptionStatus) IsTrialing() bool {
	return s.Status == StatusTrialing
}

// IsPastDue returns true if the subscription has a failed payment.
func (s *SubscriptionStatus) IsPastDue() bool {
	return s.Status == StatusPastDue
}

// IsCanceled returns true if the subscription has been canceled.
func (s *SubscriptionStatus) IsCanceled() bool {
	return s.Status == StatusCanceled
}

// Subscription status constants.
// These map to common status values from payment providers.
const (
	StatusActive   = "active"
	StatusTrialing = "trialing"
	StatusPastDue  = "past_due"
	StatusCanceled = "canceled"
	StatusUnpaid   = "unpaid"
	StatusNone     = "none" // No subscription exists
)

// IsActiveStatus returns true if the given status represents an active subscription.
// Active statuses allow access to protected features.
func IsActiveStatus(status string) bool {
	switch status {
	case StatusActive, StatusTrialing:
		return true
	default:
		return false
	}
}
