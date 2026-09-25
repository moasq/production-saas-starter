package stytch

import (
	"context"
	"encoding/json"
	"github.com/moasq/go-b2b-starter/internal/modules/auth"
	"github.com/moasq/go-b2b-starter/internal/platform/cache"
	"github.com/moasq/go-b2b-starter/internal/platform/logger"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDisabledProviderNeverAcceptsToken(t *testing.T) {
	for _, token := range []string{"", "anything", "eyJhbGciOiJub25lIn0.e30."} {
		identity, err := (DisabledAuthProvider{}).VerifyToken(context.Background(), token)
		if err == nil || identity != nil {
			t.Fatal("disabled provider authenticated a session")
		}
	}
}
func TestProviderSessionAndPermissionsAreAuthoritative(t *testing.T) {
	policyRequests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/b2b/sessions/authenticate":
			json.NewEncoder(w).Encode(map[string]any{"status_code": 200, "member": map[string]any{"member_id": "member-1", "organization_id": "org-1", "email_address": "person@example.com", "email_address_verified": true}, "member_session": map[string]any{"member_id": "member-1", "organization_id": "org-1", "expires_at": time.Now().Add(time.Hour).Format(time.RFC3339), "roles": []string{"admin"}}})
		case "/v1/b2b/rbac/policy":
			policyRequests++
			json.NewEncoder(w).Encode(map[string]any{"status_code": 200, "policy": map[string]any{"roles": []any{map[string]any{"role_id": "admin", "permissions": []any{map[string]any{"resource_id": "org", "actions": []string{"view"}}}}}}})
		default:
			t.Errorf("unexpected provider path %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	adapter, err := NewStytchAuthAdapter(&Config{ProjectID: "project-test-example", Secret: "secret-example", BaseURL: server.URL, APITimeout: time.Second}, cache.New(), logger.New())
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		identity, err := adapter.VerifyToken(context.Background(), "session")
		if err != nil {
			t.Fatal(err)
		}
		if identity.HasPermission(auth.PermOrgManage) {
			t.Fatal("admin received permission absent from provider")
		}
		if !identity.HasPermission(auth.PermOrgView) {
			t.Fatal("provider view permission lost")
		}
	}
	if policyRequests != 1 {
		t.Fatalf("policy cache requests=%d, want 1", policyRequests)
	}
}
