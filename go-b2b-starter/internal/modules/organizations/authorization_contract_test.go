package organizations

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/moasq/go-b2b-starter/internal/modules/auth"
	"github.com/moasq/go-b2b-starter/internal/platform/betterauth"
)

type contractScope struct{ unavailable bool }

func (s contractScope) Resolve(ctx context.Context, identity *auth.Identity) (context.Context, *auth.RequestContext, error) {
	if s.unavailable {
		return ctx, nil, errors.New("private database failure")
	}
	return scopeFixture{}.Resolve(ctx, identity)
}

func TestOrganizationMutationAuthorizesBeforeValidationAndSideEffects(t *testing.T) {
	t.Setenv("APP_BASE_URL", "https://app.example")
	for _, tc := range []struct {
		name, body                    string
		noCookie, expired, mismatched bool
		noPermission, unavailable     bool
		bridgeStatus, want, mutations int
	}{
		{name: "missing identity", body: `{"name":"Company"}`, noCookie: true, want: 401},
		{name: "expired identity", body: `{"name":"Company"}`, expired: true, want: 401},
		{name: "mismatched tenant", body: `{"name":"Company"}`, mismatched: true, want: 401},
		{name: "admin label without permission", body: `{"name":"Company"}`, noPermission: true, want: 403},
		{name: "authorization before validation", body: `{`, noPermission: true, want: 403},
		{name: "tenant resolution failure", body: `{"name":"Company"}`, unavailable: true, want: 503},
		{name: "malformed authorized input", body: `{`, want: 400},
		{name: "empty name", body: `{"name":""}`, want: 400},
		{name: "oversized name", body: `{"name":"` + strings.Repeat("x", 256) + `"}`, want: 400},
		{name: "provider failure", body: `{"name":"Company"}`, bridgeStatus: 500, want: 502, mutations: 1},
		{name: "verified mutation", body: `{"name":"Company","organization_id":"forged"}`, bridgeStatus: 200, want: 200, mutations: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var mutations atomic.Int32
			bridge := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/internal/auth/session" {
					expires := time.Now().Add(time.Hour)
					if tc.expired {
						expires = time.Now().Add(-time.Hour)
					}
					organization := "verified-org"
					if tc.mismatched {
						organization = "another-org"
					}
					permissions := []string{"org:view", "org:manage"}
					if tc.noPermission {
						permissions = []string{"org:view"}
					}
					json.NewEncoder(w).Encode(map[string]any{
						"user_id": "user", "member_id": "member", "email": "test@example.test", "email_verified": true,
						"organization_id": "verified-org", "organization": map[string]string{"id": organization},
						"roles": []string{"admin"}, "permissions": permissions, "expires_at": expires,
					})
					return
				}
				mutations.Add(1)
				if r.URL.Path != "/internal/auth/organization-update" || r.Header.Get("Cookie") != "better-auth.session_token=opaque" {
					t.Error("mutation did not use the expected operation and verified session")
				}
				var payload map[string]any
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || len(payload) != 1 || payload["name"] != "Company" {
					t.Errorf("unexpected mutation payload: %v (%v)", payload, err)
				}
				w.WriteHeader(tc.bridgeStatus)
				json.NewEncoder(w).Encode(map[string]any{"success": true, "data": map[string]string{"name": "Company"}})
			}))
			defer bridge.Close()
			client, err := betterauth.New(bridge.URL, strings.Repeat("s", 32))
			if err != nil {
				t.Fatal(err)
			}
			router := gin.New()
			NewRoutes(client).Routes(router.Group("/api"), middlewareResolver{auth.NewMiddleware(client, contractScope{tc.unavailable})})
			request := httptest.NewRequest(http.MethodPut, "https://app.example/api/organizations", strings.NewReader(tc.body))
			request.Header.Set("Origin", "https://app.example")
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("X-Organization-ID", "forged")
			if !tc.noCookie {
				request.Header.Set("Cookie", "better-auth.session_token=opaque")
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != tc.want || int(mutations.Load()) != tc.mutations {
				t.Fatalf("status=%d mutations=%d; want status=%d mutations=%d", response.Code, mutations.Load(), tc.want, tc.mutations)
			}
			if strings.Contains(response.Body.String(), "private database failure") {
				t.Fatal("internal database failure exposed to the client")
			}
			var result struct {
				Success bool   `json:"success"`
				Error   string `json:"error"`
				Data    struct {
					Name string `json:"name"`
				} `json:"data"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
				t.Fatalf("response is not the JSON contract: %v", err)
			}
			if tc.want == 200 && (!result.Success || result.Data.Name != "Company") || tc.want != 200 && result.Error == "" {
				t.Fatalf("unexpected response contract: %s", response.Body.String())
			}
		})
	}
}
