package polar

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/moasq/go-b2b-starter/internal/modules/billing/app/services"
	"github.com/moasq/go-b2b-starter/internal/modules/billing/domain"
	platform "github.com/moasq/go-b2b-starter/internal/platform/polar"
)

// Fixtures use the documented Polar wire shape, including external_customer_id.
func TestCheckoutAndCustomerWireContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/checkouts/checkout-id":
			_, _ = w.Write([]byte(`{"id":"checkout-id","status":"succeeded","external_customer_id":"org-owned","product_id":"product-app","customer_id":"polar-internal-id"}`))
		case "/v1/customers/external/org-owned/state":
			_, _ = w.Write([]byte(`{"external_id":"org-owned","active_subscriptions":[{"id":"subscription-id","status":"active","product_id":"product-app","current_period_end":"2099-01-01T00:00:00Z","cancel_at_period_end":true}]}`))
		default:
			w.WriteHeader(404)
			_, _ = w.Write([]byte(`{"error":"ResourceNotFound","detail":"Customer does not exist"}`))
		}
	}))
	defer server.Close()
	client, err := platform.NewClient(&platform.Config{Enabled: true, ProductID: "product-app", AccessToken: "test-only", BaseURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	adapter := NewPolarAdapter(client)
	checkout, err := adapter.GetCheckoutSession(context.Background(), "checkout-id")
	if err != nil || checkout.CustomerExternalID != "org-owned" || checkout.ProductID != "product-app" {
		t.Fatalf("checkout: %#v %v", checkout, err)
	}
	state, err := adapter.GetCustomerState(context.Background(), "org-owned")
	if err != nil || state.ExternalID != "org-owned" || len(state.ActiveSubscriptions) != 1 || !state.ActiveSubscriptions[0].CancelAtPeriodEnd {
		t.Fatalf("state: %#v %v", state, err)
	}
	missing, err := adapter.GetCustomerState(context.Background(), "org-new")
	if err != nil || missing.ExternalID != "org-new" || len(missing.ActiveSubscriptions) != 0 {
		t.Fatalf("new customer: %#v %v", missing, err)
	}
}

func TestMalformedCustomerStateIsUnavailableInsteadOfFreeOrPaid(t *testing.T) {
	for _, body := range []string{
		`null`,
		`{}`,
		`{"external_id":"org-owned"}`,
		`{"external_id":"org-owned","active_subscriptions":null}`,
		`{"external_id":"org-owned","active_subscriptions":[null]}`,
		`{"external_id":"org-owned","active_subscriptions":[{"id":"sub","status":"active","product_id":"product-app"}]}`,
		`{"external_id":"org-owned","active_subscriptions":[{"id":"sub","status":"active","product_id":"product-app","current_period_end":null}]}`,
		`{"external_id":"org-owned","active_subscriptions":[{"id":"sub","status":"active","product_id":"product-app","current_period_end":"invalid"}]}`,
	} {
		t.Run(body, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(body)) }))
			defer server.Close()
			adapter := testAdapter(t, server.URL)
			state, err := services.NewBillingService(organizationStub{}, adapter).GetBillingStatus(context.Background(), 1)
			if state != nil || !errors.Is(err, platform.ErrInvalidResponse) {
				t.Fatalf("malformed state accepted: %#v, %v", state, err)
			}
		})
	}
}

type organizationStub struct{}

func (organizationStub) GetExternalCustomerID(context.Context, int32) (string, error) {
	return "org-owned", nil
}

func testAdapter(t *testing.T, baseURL string) domain.BillingProvider {
	t.Helper()
	client, err := platform.NewClient(&platform.Config{Enabled: true, ProductID: "product-app", AccessToken: "test-only", BaseURL: baseURL})
	if err != nil {
		t.Fatal(err)
	}
	return NewPolarAdapter(client)
}

func TestCheckoutHTTPContractEnforcesTenantProductAndCompletedPayment(t *testing.T) {
	for _, tc := range []struct {
		name     string
		checkout string
		wantErr  error
	}{
		{"owned", `{"status":"succeeded","external_customer_id":"org-owned","product_id":"product-app"}`, nil},
		{"another organization", `{"status":"succeeded","external_customer_id":"org-other","product_id":"product-app"}`, domain.ErrCheckoutOwnership},
		{"another product", `{"status":"succeeded","external_customer_id":"org-owned","product_id":"another-product"}`, domain.ErrCheckoutOwnership},
		{"undocumented customer fallback", `{"status":"succeeded","customer":{"external_id":"org-owned"},"product_id":"product-app"}`, domain.ErrCheckoutOwnership},
		{"null customer identity", `{"status":"succeeded","external_customer_id":null,"product_id":"product-app"}`, domain.ErrCheckoutOwnership},
		{"confirmed is not paid", `{"status":"confirmed","external_customer_id":"org-owned","product_id":"product-app"}`, domain.ErrPaymentPending},
		{"unknown status", `{"status":"something-new","external_customer_id":"org-owned","product_id":"product-app"}`, platform.ErrInvalidResponse},
		{"null", `null`, platform.ErrInvalidResponse},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stateReads := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/v1/checkouts/checkout-id":
					_, _ = w.Write([]byte(tc.checkout))
				case "/v1/customers/external/org-owned/state":
					stateReads++
					_, _ = w.Write([]byte(`{"external_id":"org-owned","active_subscriptions":[{"id":"subscription-id","status":"active","product_id":"product-app","current_period_end":"2099-01-01T00:00:00Z","cancel_at_period_end":false}]}`))
				default:
					t.Errorf("unexpected provider path: %s", r.URL.Path)
					w.WriteHeader(500)
				}
			}))
			defer server.Close()
			service := services.NewBillingService(organizationStub{}, testAdapter(t, server.URL))
			status, err := service.VerifyPaymentFromCheckout(context.Background(), 1, "checkout-id")
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("got %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr != nil {
				if status != nil || stateReads != 0 {
					t.Fatalf("rejected checkout read customer state: %#v, %d", status, stateReads)
				}
			} else if status == nil || !status.HasActiveSubscription || stateReads != 1 {
				t.Fatalf("valid checkout did not read current subscription: %#v, %d", status, stateReads)
			}
		})
	}
}
