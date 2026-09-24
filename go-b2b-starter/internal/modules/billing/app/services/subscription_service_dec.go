package services

import (
	"context"
	"fmt"
	"time"

	"github.com/moasq/go-b2b-starter/internal/modules/billing/domain"
	platform "github.com/moasq/go-b2b-starter/internal/platform/polar"
)

type BillingService interface {
	GetBillingStatus(context.Context, int32) (*domain.BillingStatus, error)
	VerifyPaymentFromCheckout(context.Context, int32, string) (*domain.BillingStatus, error)
}

type billingService struct {
	orgAdapter domain.OrganizationAdapter
	provider   domain.BillingProvider
}

func NewBillingService(org domain.OrganizationAdapter, provider domain.BillingProvider) BillingService {
	return &billingService{orgAdapter: org, provider: provider}
}

func (s *billingService) GetBillingStatus(ctx context.Context, orgID int32) (*domain.BillingStatus, error) {
	result := &domain.BillingStatus{OrganizationID: orgID, BillingEnabled: s.provider.Enabled(), CheckedAt: time.Now().UTC()}
	if !result.BillingEnabled {
		result.Reason = "Billing is disabled"
		return result, nil
	}
	externalID, err := s.orgAdapter.GetStytchOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	state, err := s.provider.GetCustomerState(ctx, externalID)
	if err != nil {
		return nil, err
	}
	if state.ExternalID != externalID {
		return nil, fmt.Errorf("billing customer identity mismatch")
	}
	result.ExternalID = externalID
	result.Reason = "No active subscription"
	for _, sub := range state.ActiveSubscriptions {
		if sub.ProductID != s.provider.ProductID() {
			continue
		}
		if sub.Status != "active" && sub.Status != "trialing" {
			continue
		}
		// Refuse expired snapshots, including a cancellation whose access end has passed.
		if sub.CurrentPeriodEnd != nil && !sub.CurrentPeriodEnd.After(result.CheckedAt) {
			continue
		}
		if sub.EndsAt != nil && !sub.EndsAt.After(result.CheckedAt) {
			continue
		}
		result.HasActiveSubscription = true
		result.SubscriptionID, result.SubscriptionStatus = sub.ID, sub.Status
		result.ProductID, result.CurrentPeriodEnd = sub.ProductID, sub.CurrentPeriodEnd
		result.CancelAtPeriodEnd = sub.CancelAtPeriodEnd
		result.Reason = "Active subscription"
		break
	}
	return result, nil
}

// Verification is read-only. Repeating a redirect cannot grant credits or change another tenant.
func (s *billingService) VerifyPaymentFromCheckout(ctx context.Context, orgID int32, sessionID string) (*domain.BillingStatus, error) {
	if !s.provider.Enabled() {
		return nil, platform.ErrDisabled
	}
	externalID, err := s.orgAdapter.GetStytchOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	checkout, err := s.provider.GetCheckoutSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	customerID := checkout.CustomerExternalID
	if customerID == "" {
		customerID = checkout.Customer.ExternalID
	}
	if customerID == "" || customerID != externalID || checkout.ProductID != s.provider.ProductID() {
		return nil, domain.ErrCheckoutOwnership
	}
	if checkout.Status != "succeeded" {
		return nil, domain.ErrPaymentPending
	}
	return s.GetBillingStatus(ctx, orgID)
}
