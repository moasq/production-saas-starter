package services

import (
	"context"

	"github.com/moasq/go-b2b-starter/internal/billing/domain"
	logger "github.com/moasq/go-b2b-starter/internal/logger/domain"
)

// BillingService handles subscription management and quota verification.
//
// This service manages the billing lifecycle with Polar.sh via event-driven webhooks.
// It does NOT expose direct API calls to Polar during request handling - instead,
// subscription state is synced via webhooks and stored locally for fast reads.
//
// Architecture:
//
//	┌───────────────┐    webhooks    ┌─────────────────┐    reads    ┌─────────────┐
//	│   Polar.sh    │ ─────────────► │  BillingService │ ──────────► │  Local DB   │
//	└───────────────┘                └─────────────────┘             └─────────────┘
//	                                          │
//	                                          ▼
//	                                 ┌─────────────────┐
//	                                 │ Paywall reads   │
//	                                 │ from local DB   │
//	                                 └─────────────────┘
type BillingService interface {
	// ProcessWebhookEvent processes a Polar webhook event and updates local database
	// Handles: subscription.created, subscription.updated, subscription.canceled, customer.updated
	ProcessWebhookEvent(ctx context.Context, eventType string, payload map[string]any) error

	// GetBillingStatus retrieves the current billing and quota status for an organization
	// This is a read-only operation from the local database
	GetBillingStatus(ctx context.Context, organizationID int32) (*domain.BillingStatus, error)

	// CheckQuotaAvailability performs a read-only check of quota availability
	// Does NOT consume quota - use ConsumeInvoiceQuota after successful processing
	// Performs database-first check with fallback to Polar API if needed
	// Returns BillingStatus indicating if invoice processing is allowed
	CheckQuotaAvailability(ctx context.Context, organizationID int32) (*domain.BillingStatus, error)

	// ConsumeInvoiceQuota explicitly consumes one invoice quota after successful processing
	// Should be called after invoice has been successfully processed
	// Can be called asynchronously in background for better performance
	// Returns updated quota status
	ConsumeInvoiceQuota(ctx context.Context, organizationID int32) (*domain.BillingStatus, error)

	// VerifyAndConsumeQuota verifies quota availability and consumes one invoice quota
	// Performs database-first check with fallback to Polar API if needed
	// Returns BillingStatus with detailed verification result
	// Automatically increments quota count on success
	// DEPRECATED: Use CheckQuotaAvailability + ConsumeInvoiceQuota pattern for better control
	VerifyAndConsumeQuota(ctx context.Context, organizationID int32) (*domain.BillingStatus, error)

	// SyncSubscriptionFromPolar forces a sync of subscription data from Polar API
	// Used as fallback when webhook data is missing or stale
	// TODO: Implement periodic background sync job for all subscriptions
	SyncSubscriptionFromPolar(ctx context.Context, organizationID int32) error

	// VerifyPaymentFromCheckout verifies a payment by checking the Polar checkout session
	// This is the primary mechanism for "Verification on Redirect" pattern
	// Called when user returns from payment page with session_id
	// Returns BillingStatus after updating database with latest subscription info
	VerifyPaymentFromCheckout(ctx context.Context, sessionID string) (*domain.BillingStatus, error)

	// RefreshSubscriptionStatus forces a sync with Polar API and returns updated status
	// This is the lazy guarding mechanism - used when DB says expired but we want
	// to double-check with the provider in case we missed a webhook
	// Returns updated BillingStatus after syncing with provider
	RefreshSubscriptionStatus(ctx context.Context, organizationID int32) (*domain.BillingStatus, error)
}

type billingService struct {
	repo            domain.SubscriptionRepository
	orgAdapter      domain.OrganizationAdapter
	billingProvider domain.BillingProvider
	logger          logger.Logger
}

func NewBillingService(
	repo domain.SubscriptionRepository,
	orgAdapter domain.OrganizationAdapter,
	billingProvider domain.BillingProvider,
	logger logger.Logger,
) BillingService {
	return &billingService{
		repo:            repo,
		orgAdapter:      orgAdapter,
		billingProvider: billingProvider,
		logger:          logger,
	}
}
