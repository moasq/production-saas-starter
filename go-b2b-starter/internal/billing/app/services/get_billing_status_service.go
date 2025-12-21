package services

import (
	"context"
	"fmt"
	"time"

	"github.com/moasq/go-b2b-starter/internal/billing/domain"
)

func (s *billingService) GetBillingStatus(ctx context.Context, organizationID int32) (*domain.BillingStatus, error) {
	// Get quota status from database
	quotaStatus, err := s.repo.GetQuotaStatus(ctx, organizationID)
	if err != nil {
		// No subscription found
		return &domain.BillingStatus{
			OrganizationID:        organizationID,
			HasActiveSubscription: false,
			CanProcessInvoices:    false,
			InvoiceCount:          0,
			Reason:                "no active subscription found",
			CheckedAt:             time.Now(),
		}, nil
	}

	// Build billing status from quota status
	return &domain.BillingStatus{
		OrganizationID:        organizationID,
		HasActiveSubscription: quotaStatus.SubscriptionStatus == "active",
		CanProcessInvoices:    quotaStatus.CanProcessInvoice,
		InvoiceCount:          quotaStatus.InvoiceCount,
		Reason:                s.buildStatusReason(quotaStatus),
		CheckedAt:             time.Now(),
	}, nil
}

func (s *billingService) buildStatusReason(status *domain.QuotaStatus) string {
	if !status.CanProcessInvoice {
		if status.SubscriptionStatus != "active" {
			return fmt.Sprintf("subscription status: %s", status.SubscriptionStatus)
		}
		return "invoice quota exceeded"
	}
	return "ok"
}
