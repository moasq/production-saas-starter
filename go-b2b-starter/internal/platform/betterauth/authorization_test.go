package betterauth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/moasq/go-b2b-starter/internal/modules/auth"
)

func TestCurrentBridgePermissionsControlProtectedRequests(t *testing.T) {
	t.Setenv("APP_BASE_URL", "https://app.example")
	type response struct {
		status int
		roles  []string
		grants []string
	}
	var current atomic.Value
	var calls atomic.Int32
	bridge := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		state := current.Load().(response)
		w.WriteHeader(state.status)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"user_id": "user-1", "member_id": "member-1", "email": "one@example.test", "email_verified": true,
			"organization_id": "org-1", "organization": map[string]string{"id": "org-1"},
			"roles": state.roles, "permissions": state.grants, "expires_at": time.Now().Add(time.Hour),
		})
	}))
	defer bridge.Close()
	client, err := New(bridge.URL, strings.Repeat("s", 32))
	if err != nil {
		t.Fatal(err)
	}
	mutations := 0
	router := gin.New()
	router.POST("/organization", auth.NewMiddleware(client, nil).RequireAuth(), auth.RequirePermissionFunc("org", "manage"), func(c *gin.Context) {
		mutations++
		c.Status(http.StatusNoContent)
	})
	for _, tc := range []struct {
		name  string
		state response
		want  int
	}{
		{"initial administrator", response{200, []string{"admin"}, []string{"org:view", "org:manage"}}, 204},
		{"restrictive policy keeps role but removes grant", response{200, []string{"admin"}, []string{"org:view"}}, 403},
		{"explicit grant restored", response{200, []string{"admin"}, []string{"org:view", "org:manage"}}, 204},
		{"membership demoted", response{200, []string{"member"}, []string{"org:view"}}, 403},
		{"no grants", response{200, []string{"admin"}, nil}, 403},
		{"unknown role with grant", response{200, []string{"custom-admin"}, []string{"org:manage"}}, 401},
		{"combined role with grant", response{200, []string{"admin,member"}, []string{"org:manage"}}, 401},
		{"multiple roles with grant", response{200, []string{"admin", "member"}, []string{"org:manage"}}, 401},
		{"missing role with grant", response{200, nil, []string{"org:manage"}}, 401},
		{"wildcard is outside application contract", response{200, []string{"admin"}, []string{"*:*"}}, 401},
		{"unknown permission", response{200, []string{"admin"}, []string{"org:manage", "billing:*"}}, 401},
		{"unavailable policy authority", response{503, []string{"admin"}, []string{"org:manage"}}, 401},
		{"revoked session", response{401, []string{"admin"}, []string{"org:manage"}}, 401},
	} {
		t.Run(tc.name, func(t *testing.T) {
			current.Store(tc.state)
			before := mutations
			req := httptest.NewRequest(http.MethodPost, "https://app.example/organization", nil)
			req.Header.Set("Cookie", "better-auth.session_token=same-session")
			req.Header.Set("Origin", "https://app.example")
			req.Header.Set("X-Role", "admin")
			req.Header.Set("X-Permissions", "org:manage")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d: %s", rec.Code, tc.want, rec.Body.String())
			}
			if tc.want != http.StatusNoContent && mutations != before {
				t.Fatal("denied request reached business mutation")
			}
		})
	}
	if calls.Load() != 13 || mutations != 2 {
		t.Fatalf("bridge calls = %d, mutations = %d; expected a fresh check per request and only two authorized writes", calls.Load(), mutations)
	}
}
