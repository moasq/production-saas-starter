package organizations

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/moasq/go-b2b-starter/internal/modules/auth"
	"github.com/moasq/go-b2b-starter/internal/platform/betterauth"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type scopeFixture struct{}

func (scopeFixture) Resolve(ctx context.Context, i *auth.Identity) (context.Context, *auth.RequestContext, error) {
	return ctx, &auth.RequestContext{Identity: i, OrganizationID: 1, AccountID: 1, ProviderOrgID: i.OrganizationID}, nil
}

type middlewareResolver struct{ m *auth.Middleware }

func (r middlewareResolver) Get(name string) gin.HandlerFunc {
	if name == "auth" {
		return r.m.RequireAuth()
	}
	if name == "org_context" {
		return r.m.RequireOrganization()
	}
	panic(name)
}

func TestMemberMutationsRequireLivePermissionAndDropClientTenant(t *testing.T) {
	t.Setenv("APP_BASE_URL", "https://app.example")
	permission := "org:view"
	operations := 0
	bridge := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/internal/auth/session" {
			json.NewEncoder(w).Encode(map[string]any{"user_id": "user", "member_id": "member", "email": "one@example.com", "email_verified": true, "organization_id": "verified-org", "organization": map[string]string{"id": "verified-org", "name": "Company", "slug": "company"}, "roles": []string{"admin"}, "permissions": []string{permission}, "expires_at": time.Now().Add(time.Hour)})
			return
		}
		operations++
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if _, ok := body["organization_id"]; ok {
			t.Fatal("client tenant forwarded")
		}
		if _, ok := body["org_id"]; ok {
			t.Fatal("client tenant forwarded")
		}
		if r.Header.Get("Cookie") != "better-auth.session_token=opaque" {
			t.Error("end-user session was omitted")
		}
		json.NewEncoder(w).Encode(map[string]any{"success": true, "data": map[string]bool{"invite_sent": true}})
	}))
	defer bridge.Close()
	client, err := betterauth.New(bridge.URL, strings.Repeat("s", 32))
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	NewRoutes(client).Routes(router.Group("/api"), middlewareResolver{auth.NewMiddleware(client, scopeFixture{})})
	for _, tc := range []struct{ method, path, body string }{
		{"POST", "/api/auth/members", `{"email":"two@example.com","name":"Two","role_slug":"member","org_id":"attacker"}`},
		{"DELETE", "/api/auth/members/member-2", ``},
		{"POST", "/api/auth/members/invitation:pending/resend-invitation", ``},
		{"PUT", "/api/auth/members/member-2", `{"role":"member","organization_id":"attacker"}`},
	} {
		for _, want := range []int{403, 200} {
			permission = "org:view"
			if want == 200 {
				permission = "org:manage"
			}
			before := operations
			req := httptest.NewRequest(tc.method, "https://app.example"+tc.path, strings.NewReader(tc.body))
			req.Header.Set("Cookie", "better-auth.session_token=opaque")
			req.Header.Set("Origin", "https://app.example")
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != want {
				t.Fatalf("%s: status=%d want=%d body=%s", tc.path, rec.Code, want, rec.Body)
			}
			if want == 403 && operations != before {
				t.Fatal("unprivileged mutation reached bridge")
			}
		}
	}
}
