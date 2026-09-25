package auth

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testProvider struct{ calls int }

func (p *testProvider) VerifyToken(context.Context, string) (*Identity, error) {
	p.calls++
	return &Identity{Roles: []Role{RoleAdmin}}, nil
}

func TestCookieAuthenticationRequiresSameOriginForWrites(t *testing.T) {
	t.Setenv("APP_BASE_URL", "")
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, method, origin, header string
		want                         int
	}{
		{"same origin", "POST", "https://app.example", "", 200},
		{"cross origin", "POST", "https://attacker.example", "", 403},
		{"missing origin", "POST", "", "", 403},
		{"safe read", "GET", "", "", 200},
		{"server bearer", "POST", "", "Bearer server-token", 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			provider := &testProvider{}
			router := gin.New()
			router.Use(NewMiddleware(provider, nil, nil, nil).RequireAuth())
			router.Any("/members", func(c *gin.Context) { c.Status(200) })
			req := httptest.NewRequest(tc.method, "https://app.example/members", nil)
			req.AddCookie(&http.Cookie{Name: "stytch_session_jwt", Value: "test-token"})
			req.Header.Set("Origin", tc.origin)
			req.Header.Set("Authorization", tc.header)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("status %d, want %d", rec.Code, tc.want)
			}
			if tc.want == 403 && provider.calls != 0 {
				t.Fatal("cross-origin write reached provider")
			}
		})
	}
}

func TestRoleNameDoesNotGrantPermissions(t *testing.T) {
	identity := &Identity{Roles: []Role{RoleAdmin}}
	if hasPermission(identity, "org", "manage") {
		t.Fatal("admin role bypassed provider policy")
	}
	identity.Permissions = []Permission{PermOrgManage}
	if !hasPermission(identity, "org", "manage") {
		t.Fatal("explicit provider permission denied")
	}
}
