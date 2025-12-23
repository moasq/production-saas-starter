package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/moasq/go-b2b-starter/internal/modules/billing/domain"
)

func (s *billingService) VerifyPaymentFromCheckout(ctx context.Context, sessionID string) (*domain.BillingStatus, error) {
	// Step 1: Get checkout session from Polar with polling
	checkoutSession, err := s.billingProvider.GetCheckoutSessionWithPolling(ctx, sessionID)
	if err != nil {
		fmt.Printf("❌ [VerifyPayment] Failed to verify checkout session %s: %v\n", sessionID, err)
		return nil, fmt.Errorf("failed to get checkout session: %w", err)
	}

	fmt.Printf("✅ [VerifyPayment] Checkout session %s verified with status: %s\n", sessionID, checkoutSession.Status)

	// Step 2: Verify checkout status
	if checkoutSession.Status != "succeeded" {
		return &domain.BillingStatus{
			HasActiveSubscription: false,
			CanProcessInvoices:    false,
			Reason:                fmt.Sprintf("checkout session status is %s (expected: succeeded)", checkoutSession.Status),
			CheckedAt:             time.Now(),
		}, nil
	}

	// Step 3: Extract customer ID (this is the Stytch org ID)
	externalCustomerID := checkoutSession.CustomerID
	if externalCustomerID == "" {
		return nil, fmt.Errorf("checkout session has no customer_id")
	}

	// Step 4: Map external customer ID to internal organization ID
	organizationID, err := s.orgAdapter.GetOrganizationIDByStytchOrgID(ctx, externalCustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to map customer ID to organization: %w", err)
	}

	// Step 5: Fetch full subscription details from Polar
	subscription, err := s.billingProvider.GetSubscription(ctx, externalCustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subscription from Polar: %w", err)
	}

	// Step 6: Upsert subscription to database
	subscription.OrganizationID = organizationID
	_, err = s.repo.UpsertSubscription(ctx, subscription)
	if err != nil {
		return nil, fmt.Errorf("failed to save subscription: %w", err)
	}

	// Step 7: Extract and upsert quota information
	invoiceCountMax := int32(0)
	if metadata, ok := subscription.Metadata["invoice_count_max"].(int32); ok {
		invoiceCountMax = metadata
	} else if val, ok := subscription.Metadata["invoice_count_max"].(string); ok {
		if count, err := strconv.ParseInt(val, 10, 32); err == nil {
			invoiceCountMax = int32(count)
		}
	}

	// Create or update quota tracking
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
		return nil, fmt.Errorf("failed to save quota: %w", err)
	}

	s.logger.Info("Payment verified from checkout session", map[string]any{
		"session_id":      sessionID,
		"organization_id": organizationID,
		"subscription_id": subscription.SubscriptionID,
		"invoice_count":   invoiceCountMax,
	})

	// Console log for verification completion
	fmt.Printf("✅ PAYMENT VERIFIED - Session: %s | Org: %d | Subscription: %s | Invoice Count: %d | Status: %s\n",
		sessionID, organizationID, subscription.SubscriptionID, invoiceCountMax, subscription.SubscriptionStatus)

	// Step 8: Return billing status
	return &domain.BillingStatus{
		OrganizationID:        organizationID,
		ExternalID:            externalCustomerID,
		HasActiveSubscription: subscription.SubscriptionStatus == "active" || subscription.SubscriptionStatus == "trialing",
		CanProcessInvoices:    (subscription.SubscriptionStatus == "active" || subscription.SubscriptionStatus == "trialing") && invoiceCountMax > 0,
		InvoiceCount:          invoiceCountMax,
		Reason:                "Payment verified successfully",
		CheckedAt:             time.Now(),
	}, nil
}
