package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/moasq/go-b2b-starter/internal/modules/billing/domain"
	platform "github.com/moasq/go-b2b-starter/internal/platform/polar"
)

type orgStub struct{}

func (orgStub) GetExternalCustomerID(context.Context, int32) (string, error) { return "org-owned", nil }

type providerStub struct {
	enabled  bool
	state    domain.CustomerState
	checkout domain.CheckoutSession
	err      error
	reads    int
}

func (p *providerStub) ProductID() string { return "product-app" }
func (p *providerStub) Enabled() bool     { return p.enabled }
func (p *providerStub) GetCustomerState(context.Context, string) (*domain.CustomerState, error) {
	p.reads++
	return &p.state, p.err
}
func (p *providerStub) GetCheckoutSession(context.Context, string) (*domain.CheckoutSession, error) {
	return &p.checkout, p.err
}
func TestCheckoutCannotAccessAnotherOrganization(t *testing.T) {
	p := &providerStub{enabled: true, checkout: domain.CheckoutSession{ProductID: "product-app", Status: "succeeded", CustomerExternalID: "org-other"}}
	_, err := NewBillingService(orgStub{}, p).VerifyPaymentFromCheckout(context.Background(), 1, "checkout")
	if !errors.Is(err, domain.ErrCheckoutOwnership) || p.reads != 0 {
		t.Fatalf("got %v, customer reads %d", err, p.reads)
	}
}
func TestCheckoutRepeatReadsCurrentStateWithoutGrantingAnything(t *testing.T) {
	p := &providerStub{enabled: true, checkout: domain.CheckoutSession{ProductID: "product-app", Status: "succeeded", CustomerExternalID: "org-owned"}, state: domain.CustomerState{ExternalID: "org-owned"}}
	service := NewBillingService(orgStub{}, p)
	for range 2 {
		status, err := service.VerifyPaymentFromCheckout(context.Background(), 1, "checkout")
		if err != nil || status.HasActiveSubscription {
			t.Fatalf("successful checkout is not itself an entitlement: %#v %v", status, err)
		}
	}
	if p.reads != 2 {
		t.Fatal("each read must query current provider state")
	}
}
func TestStatusRequiresUnexpiredActiveSubscription(t *testing.T) {
	past, future := time.Now().Add(-time.Hour), time.Now().Add(time.Hour)
	for _, tc := range []struct {
		name, status string
		end, ends    *time.Time
		want         bool
	}{
		{"active", "active", &future, nil, true}, {"trial", "trialing", &future, nil, true},
		{"expired", "active", &past, nil, false}, {"canceled", "canceled", &future, nil, false},
		{"ended", "active", &future, &past, false}, {"no period end", "active", nil, nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &providerStub{enabled: true, state: domain.CustomerState{ExternalID: "org-owned", ActiveSubscriptions: []domain.Subscription{{ProductID: "product-app", Status: tc.status, CurrentPeriodEnd: tc.end, EndsAt: tc.ends}}}}
			got, err := NewBillingService(orgStub{}, p).GetBillingStatus(context.Background(), 1)
			if err != nil || got.HasActiveSubscription != tc.want {
				t.Fatalf("got %#v, %v", got, err)
			}
		})
	}
}
func TestDisabledBillingDoesNotContactProvider(t *testing.T) {
	p := &providerStub{}
	service := NewBillingService(orgStub{}, p)
	result, err := service.GetBillingStatus(context.Background(), 1)
	if err != nil || result.BillingEnabled || result.HasActiveSubscription || p.reads != 0 {
		t.Fatalf("got %#v %v", result, err)
	}
	_, err = service.VerifyPaymentFromCheckout(context.Background(), 1, "checkout")
	if !errors.Is(err, platform.ErrDisabled) {
		t.Fatal(err)
	}
}
func TestProviderFailureIsNotFreeOrPaidStatus(t *testing.T) {
	failure := errors.New("provider unavailable")
	p := &providerStub{enabled: true, err: failure}
	result, err := NewBillingService(orgStub{}, p).GetBillingStatus(context.Background(), 1)
	if result != nil || !errors.Is(err, failure) {
		t.Fatalf("got %#v %v", result, err)
	}
}

func TestUnrelatedPolarProductDoesNotActivateThisApplication(t *testing.T) {
	p := &providerStub{enabled: true, state: domain.CustomerState{ExternalID: "org-owned", ActiveSubscriptions: []domain.Subscription{{Status: "active", ProductID: "another-app"}}}}
	status, err := NewBillingService(orgStub{}, p).GetBillingStatus(context.Background(), 1)
	if err != nil || status.HasActiveSubscription {
		t.Fatalf("unexpected entitlement: %#v %v", status, err)
	}
	p.checkout = domain.CheckoutSession{Status: "succeeded", CustomerExternalID: "org-owned", ProductID: "another-app"}
	_, err = NewBillingService(orgStub{}, p).VerifyPaymentFromCheckout(context.Background(), 1, "checkout")
	if !errors.Is(err, domain.ErrCheckoutOwnership) {
		t.Fatal("unrelated product checkout accepted")
	}
}
