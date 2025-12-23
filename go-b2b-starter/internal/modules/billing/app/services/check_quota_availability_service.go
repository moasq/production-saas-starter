package services

import (
	"context"
	"fmt"
	"time"

	"github.com/moasq/go-b2b-starter/internal/modules/billing/domain"
)

// CheckQuotaAvailability performs a read-only verification of quota availability
// This method does NOT consume quota - it only checks if processing is allowed
// Use ConsumeInvoiceQuota after successful invoice processing to actually decrement the quota
func (s *billingService) CheckQuotaAvailability(ctx context.Context, organizationID int32) (*domain.BillingStatus, error) {
	// Step 1: Check database quota status (read-only)
	quotaStatus, err := s.repo.GetQuotaStatus(ctx, organizationID)
	if err != nil {
		return &domain.BillingStatus{
			OrganizationID:        organizationID,
			HasActiveSubscription: false,
			CanProcessInvoices:    false,
			Reason:                "no active subscription",
			CheckedAt:             time.Now(),
		}, domain.ErrSubscriptionNotFound
	}

	// Step 2: Check if we need fallback API verification
	needsFallback := s.needsFallbackVerification(quotaStatus)
	if needsFallback {
		s.logger.Info("Quota near limit or stale, performing fallback API verification", map[string]any{
			"organization_id": organizationID,
			"invoice_count":   quotaStatus.InvoiceCount,
		})

		// Sync from Polar and re-check
		if err := s.SyncSubscriptionFromPolar(ctx, organizationID); err != nil {
			s.logger.Error("Fallback sync failed, using database data", map[string]any{
				"organization_id": organizationID,
				"error":           err.Error(),
			})
		} else {
			// Re-fetch quota status after sync
			quotaStatus, err = s.repo.GetQuotaStatus(ctx, organizationID)
			if err != nil {
				return nil, fmt.Errorf("failed to get quota after sync: %w", err)
			}
		}
	}

	// Step 3: Verify quota is available (NO consumption here)
	if !quotaStatus.CanProcessInvoice {
		return &domain.BillingStatus{
			OrganizationID:        organizationID,
			HasActiveSubscription: quotaStatus.SubscriptionStatus == "active",
			CanProcessInvoices:    false,
			InvoiceCount:          quotaStatus.InvoiceCount,
			Reason:                "quota exceeded or subscription inactive",
			CheckedAt:             time.Now(),
		}, domain.ErrQuotaExceeded
	}

	// Step 4: Return success status (quota NOT consumed yet)
	return &domain.BillingStatus{
		OrganizationID:        organizationID,
		HasActiveSubscription: true,
		CanProcessInvoices:    true,
		InvoiceCount:          quotaStatus.InvoiceCount, // Current count, NOT decremented
		Reason:                "quota available",
		CheckedAt:             time.Now(),
	}, nil
}
