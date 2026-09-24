package polar

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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
