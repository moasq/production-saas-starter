// Package adapters provides adapter implementations for external interfaces.
package adapters

import (
	"context"

	"github.com/moasq/go-b2b-starter/internal/billing/app/services"
	"github.com/moasq/go-b2b-starter/internal/paywall"
)

// StatusProviderAdapter adapts the BillingService to the SubscriptionStatusProvider interface.
//
// This adapter allows the paywall middleware to check subscription status
// without depending directly on the billing service implementation.
// Communication is event-driven: Polar webhooks → billing module → local DB → paywall reads.
type StatusProviderAdapter struct {
	service services.BillingService
}

func NewStatusProviderAdapter(service services.BillingService) paywall.SubscriptionStatusProvider {
	return &StatusProviderAdapter{service: service}
}

// GetSubscriptionStatus implements paywall.SubscriptionStatusProvider.
//
// It delegates to the BillingService.GetBillingStatus method and converts
// the BillingStatus to a SubscriptionStatus for the middleware to use.
func (a *StatusProviderAdapter) GetSubscriptionStatus(ctx context.Context, organizationID int32) (*paywall.SubscriptionStatus, error) {
	billingStatus, err := a.service.GetBillingStatus(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	// Map BillingStatus to SubscriptionStatus
	status := &paywall.SubscriptionStatus{
		OrganizationID: billingStatus.OrganizationID,
		IsActive:       billingStatus.HasActiveSubscription,
		Reason:         billingStatus.Reason,
	}

	// Determine status string from reason
	if billingStatus.HasActiveSubscription {
		status.Status = paywall.StatusActive
	} else if billingStatus.Reason == "no active subscription found" {
		status.Status = paywall.StatusNone
	} else {
		// Parse status from reason if available, otherwise default to inactive
		status.Status = parseStatusFromReason(billingStatus.Reason)
	}

	return status, nil
}

// RefreshSubscriptionStatus implements paywall.SubscriptionStatusProvider.
//
// It forces a sync with the payment provider API and returns the updated status.
// This is the lazy guarding mechanism - used when DB says expired but we want
// to double-check with the provider in case we missed a webhook.
func (a *StatusProviderAdapter) RefreshSubscriptionStatus(ctx context.Context, organizationID int32) (*paywall.SubscriptionStatus, error) {
	// Delegate to the BillingService.RefreshSubscriptionStatus method
	billingStatus, err := a.service.RefreshSubscriptionStatus(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	// Map BillingStatus to SubscriptionStatus
	status := &paywall.SubscriptionStatus{
		OrganizationID: billingStatus.OrganizationID,
		IsActive:       billingStatus.HasActiveSubscription,
		Reason:         billingStatus.Reason,
	}

	// Determine status string from reason
	if billingStatus.HasActiveSubscription {
		status.Status = paywall.StatusActive
	} else if billingStatus.Reason == "no active subscription found" {
		status.Status = paywall.StatusNone
	} else {
		// Parse status from reason if available, otherwise default to inactive
		status.Status = parseStatusFromReason(billingStatus.Reason)
	}

	return status, nil
}

// parseStatusFromReason attempts to extract a subscription status from the reason string.
func parseStatusFromReason(reason string) string {
	// Check for common status patterns in reason
	switch {
	case containsStatus(reason, "past_due"):
		return paywall.StatusPastDue
	case containsStatus(reason, "canceled"):
		return paywall.StatusCanceled
	case containsStatus(reason, "unpaid"):
		return paywall.StatusUnpaid
	case containsStatus(reason, "trialing"):
		return paywall.StatusTrialing
	default:
		return paywall.StatusNone
	}
}

// containsStatus checks if the reason contains a specific status.
func containsStatus(reason, status string) bool {
	return len(reason) >= len(status) && contains(reason, status)
}

// contains is a simple substring check.
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
