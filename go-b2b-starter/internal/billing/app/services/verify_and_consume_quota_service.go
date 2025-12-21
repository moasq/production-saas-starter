package services

import (
	"context"
	"fmt"
	"time"

	"github.com/moasq/go-b2b-starter/internal/billing/domain"
)

func (s *billingService) VerifyAndConsumeQuota(ctx context.Context, organizationID int32) (*domain.BillingStatus, error) {
	// Step 1: Check database quota status
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

	// Step 3: Verify quota is available
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

	// Step 4: Decrement quota count (consume one invoice)
	_, err = s.repo.DecrementInvoiceCount(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to decrement invoice count: %w", err)
	}

	// Step 5: Return success status
	return &domain.BillingStatus{
		OrganizationID:        organizationID,
		HasActiveSubscription: true,
		CanProcessInvoices:    true,
		InvoiceCount:          quotaStatus.InvoiceCount - 1, // Already decremented
		Reason:                "quota verified and consumed",
		CheckedAt:             time.Now(),
	}, nil
}

func (s *billingService) needsFallbackVerification(status *domain.QuotaStatus) bool {
	// Perform fallback if:
	// 1. Very few invoices remaining (< 10)
	// 2. Subscription is inactive but we're checking

	return status.InvoiceCount < 10 || status.SubscriptionStatus != "active"
}
