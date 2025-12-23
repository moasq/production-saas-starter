package services

import (
	"context"
	"fmt"
	"time"

	"github.com/moasq/go-b2b-starter/internal/modules/billing/domain"
)

// ConsumeInvoiceQuota explicitly consumes one invoice quota after successful processing
// This should be called after the invoice has been successfully processed
// Can be safely called in a background goroutine for better performance
func (s *billingService) ConsumeInvoiceQuota(ctx context.Context, organizationID int32) (*domain.BillingStatus, error) {
	s.logger.Info("Consuming invoice quota for organization", map[string]any{
		"organization_id": organizationID,
	})

	// Step 1: Get current quota status before consumption
	quotaStatus, err := s.repo.GetQuotaStatus(ctx, organizationID)
	if err != nil {
		s.logger.Error("Failed to get quota status before consumption", map[string]any{
			"organization_id": organizationID,
			"error":           err.Error(),
		})
		return nil, fmt.Errorf("failed to get quota status: %w", err)
	}

	// Step 2: Decrement quota count (atomic database operation)
	updatedQuota, err := s.repo.DecrementInvoiceCount(ctx, organizationID)
	if err != nil {
		s.logger.Error("Failed to decrement invoice count", map[string]any{
			"organization_id": organizationID,
			"error":           err.Error(),
		})
		return nil, fmt.Errorf("failed to decrement invoice count: %w", err)
	}

	s.logger.Info("Successfully consumed invoice quota locally", map[string]any{
		"organization_id":    organizationID,
		"previous_count":     quotaStatus.InvoiceCount,
		"new_count":          updatedQuota.InvoiceCount,
		"remaining_invoices": updatedQuota.InvoiceCount,
	})

	// Step 3: Ingest meter event to Polar to consume credits (best-effort)
	// This notifies Polar about the invoice processing usage
	// Local tracking is maintained for fast quota checks, Polar tracks actual billing
	go s.ingestMeterEventToPolar(context.Background(), organizationID)

	// Step 4: Return updated billing status
	return &domain.BillingStatus{
		OrganizationID:        organizationID,
		HasActiveSubscription: quotaStatus.SubscriptionStatus == "active",
		CanProcessInvoices:    updatedQuota.InvoiceCount > 0,
		InvoiceCount:          updatedQuota.InvoiceCount,
		Reason:                "quota consumed successfully",
		CheckedAt:             time.Now(),
	}, nil
}

// ingestMeterEventToPolar ingests a meter event to Polar for usage-based billing
// This runs in a background goroutine and uses best-effort approach
// Failures are logged but don't affect the main operation since local tracking is maintained
func (s *billingService) ingestMeterEventToPolar(ctx context.Context, organizationID int32) {
	// Use background context with timeout (independent of request context)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Get organization's external customer ID (Stytch org ID)
	externalID, err := s.orgAdapter.GetStytchOrgID(ctx, organizationID)
	if err != nil {
		s.logger.Error("Failed to get external customer ID for Polar meter event", map[string]any{
			"organization_id": organizationID,
			"error":           err.Error(),
		})
		return
	}

	// Ingest meter event to Polar
	// Meter: "Invoice Processing" (configured in Polar dashboard)
	// Filter: name equals "invoice.processed"
	// Amount: 1 (one invoice processed)
	meterSlug := invoicesProcessedMeterSlug // Event name MUST match meter filter exactly (with dot)
	if err := s.billingProvider.IngestMeterEvent(ctx, externalID, meterSlug, 1); err != nil {
		s.logger.Error("Failed to ingest meter event to Polar", map[string]any{
			"organization_id": organizationID,
			"external_id":     externalID,
			"meter_slug":      meterSlug,
			"error":           err.Error(),
		})
		return
	}

	// Log success
	s.logger.Info("Successfully ingested event to Polar", map[string]any{
		"organization_id": organizationID,
		"external_id":     externalID,
		"event_name":      meterSlug,
		"amount":          1,
	})
}
