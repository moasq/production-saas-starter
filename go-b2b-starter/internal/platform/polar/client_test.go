package polar

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConfigUsesEnvironmentWithoutFile(t *testing.T) {
	t.Setenv("BILLING_ENABLED", "true")
	t.Setenv("POLAR_PRODUCT_ID", "product-app")
	t.Setenv("POLAR_ACCESS_TOKEN", "test-only-token")
	t.Setenv("POLAR_ENVIRONMENT", "sandbox")
	cfg, err := LoadConfig()
	if err != nil || !cfg.Enabled || cfg.BaseURL != "https://sandbox-api.polar.sh" {
		t.Fatalf("%#v %v", cfg, err)
	}
	t.Setenv("POLAR_ENVIRONMENT", "production")
	cfg, err = LoadConfig()
	if err != nil || cfg.BaseURL != "https://api.polar.sh" {
		t.Fatalf("%#v %v", cfg, err)
	}
	t.Setenv("POLAR_ACCESS_TOKEN", "")
	if _, err = LoadConfig(); err == nil {
		t.Fatal("enabled billing must require credentials")
	}
	t.Setenv("BILLING_ENABLED", "false")
	if _, err = LoadConfig(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("POLAR_ENVIRONMENT", "typo")
	if _, err = LoadConfig(); err == nil {
		t.Fatal("must reject unknown provider environment")
	}
}
func TestClientHandlesNotFoundAndRedactsProviderBodies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-only-token" {
			t.Error("missing authentication")
		}
		if r.URL.Path == "/missing" {
			w.WriteHeader(404)
		} else {
			w.WriteHeader(500)
		}
		_, _ = w.Write([]byte("sensitive-provider-body"))
	}))
	defer server.Close()
	client, err := NewClient(&Config{Enabled: true, ProductID: "product-app", AccessToken: "test-only-token", BaseURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	if err = client.GetJSON(context.Background(), "/missing", &struct{}{}); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	err = client.GetJSON(context.Background(), "/error", &struct{}{})
	if err == nil || strings.Contains(err.Error(), "sensitive-provider-body") {
		t.Fatalf("unsafe error: %v", err)
	}
}
