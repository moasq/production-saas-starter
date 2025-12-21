package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/moasq/go-b2b-starter/internal/billing/domain"
)

func (s *billingService) SyncSubscriptionFromPolar(ctx context.Context, organizationID int32) error {
	// Get organization's external customer ID
	externalID, err := s.orgAdapter.GetStytchOrgID(ctx, organizationID)
	if err != nil {
		return fmt.Errorf("failed to get organization external ID: %w", err)
	}

	// Fetch subscription from Polar
	subscription, err := s.billingProvider.GetSubscription(ctx, externalID)
	if err != nil {
		return fmt.Errorf("failed to fetch subscription from Polar: %w", err)
	}

	// Upsert subscription to database
	subscription.OrganizationID = organizationID
	_, err = s.repo.UpsertSubscription(ctx, subscription)
	if err != nil {
		return fmt.Errorf("failed to save subscription: %w", err)
	}

	// Extract and upsert quota information
	invoiceCountMax := int32(0)
	if metadata, ok := subscription.Metadata["invoice_count_max"].(int32); ok {
		invoiceCountMax = metadata
	} else if val, ok := subscription.Metadata["invoice_count_max"].(string); ok {
		if count, err := strconv.ParseInt(val, 10, 32); err == nil {
			invoiceCountMax = int32(count)
		}
	}

	// Create or update quota tracking with synced data
	quota := &domain.QuotaTracking{
		OrganizationID: organizationID,
		InvoiceCount:   invoiceCountMax,
		PeriodStart:    subscription.CurrentPeriodStart,
		PeriodEnd:      subscription.CurrentPeriodEnd,
		LastSyncedAt:   &time.Time{},
	}
	*quota.LastSyncedAt = time.Now()

	_, err = s.repo.UpsertQuota(ctx, quota)
	if err != nil {
		return fmt.Errorf("failed to save quota: %w", err)
	}

	s.logger.Info("Synced subscription and quota from Polar", map[string]any{
		"organization_id": organizationID,
		"subscription_id": subscription.SubscriptionID,
		"invoice_count":   invoiceCountMax,
		"synced_at":       quota.LastSyncedAt,
	})

	// Console log for sync completion
	fmt.Printf("🔄 SYNC COMPLETED - Org: %d | Subscription: %s | Invoice Count: %d | Status: %s | Synced at: %s\n",
		organizationID, subscription.SubscriptionID, invoiceCountMax, subscription.SubscriptionStatus, quota.LastSyncedAt.Format(time.RFC3339))

	return nil
}
