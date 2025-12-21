package services

import (
	"context"
	"fmt"
	"time"

	"github.com/moasq/go-b2b-starter/internal/billing/domain"
)

// RefreshSubscriptionStatus forces a sync with Polar API and returns updated status.
// This is the lazy guarding mechanism - used when DB says expired but we want
// to double-check with the provider in case we missed a webhook.
func (s *billingService) RefreshSubscriptionStatus(ctx context.Context, organizationID int32) (*domain.BillingStatus, error) {
	// Step 1: Check if subscription exists in database
	_, err := s.repo.GetSubscriptionByOrgID(ctx, organizationID)
	if err != nil {
		// No subscription exists - don't call Polar API
		s.logger.Info("No subscription found for refresh", map[string]any{
			"organization_id": organizationID,
		})

		return &domain.BillingStatus{
			OrganizationID:        organizationID,
			HasActiveSubscription: false,
			CanProcessInvoices:    false,
			InvoiceCount:          0,
			Reason:                "no active subscription found",
			CheckedAt:             time.Now(),
		}, nil
	}

	// Step 2: Sync subscription from Polar API
	if err := s.SyncSubscriptionFromPolar(ctx, organizationID); err != nil {
		// Sync failed - return error
		return nil, fmt.Errorf("failed to refresh subscription from Polar: %w", err)
	}

	// Step 3: Get fresh billing status from database (after sync)
	billingStatus, err := s.GetBillingStatus(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get billing status after refresh: %w", err)
	}

	s.logger.Info("Subscription status refreshed", map[string]any{
		"organization_id":         organizationID,
		"has_active_subscription": billingStatus.HasActiveSubscription,
		"invoice_count":           billingStatus.InvoiceCount,
	})

	// Console log for refresh completion
	fmt.Printf("🔄 SUBSCRIPTION REFRESHED - Org: %d | Active: %v | Invoice Count: %d | Reason: %s\n",
		organizationID, billingStatus.HasActiveSubscription, billingStatus.InvoiceCount, billingStatus.Reason)

	return billingStatus, nil
}
