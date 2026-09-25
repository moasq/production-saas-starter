package auth

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testProvider struct {
	calls   int
	revoked bool
}

func (p *testProvider) VerifySession(context.Context, string) (*Identity, error) {
	p.calls++
	if p.revoked {
		return nil, ErrUnauthorized
	}
	return &Identity{Roles: []Role{RoleAdmin}}, nil
}
func TestCookieAuthenticationRequiresSameOriginForWrites(t *testing.T) {
	t.Setenv("APP_BASE_URL", "https://app.example")
	for _, tc := range []struct {
		name, method, origin, authorization string
		want                                int
	}{
		{"same origin", "POST", "https://app.example", "", 200},
		{"cross origin", "POST", "https://attacker.example", "", 403},
		{"missing origin", "POST", "", "", 403},
		{"safe read", "GET", "", "", 200},
		{"forged bearer does not bypass CSRF", "POST", "", "Bearer client-asserted", 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := &testProvider{}
			router := gin.New()
			router.Use(NewMiddleware(p, nil).RequireAuth())
			router.Any("/members", func(c *gin.Context) { c.Status(200) })
			req := httptest.NewRequest(tc.method, "https://app.example/members", nil)
			req.AddCookie(&http.Cookie{Name: "better-auth.session_token", Value: "opaque"})
			req.Header.Set("Origin", tc.origin)
			req.Header.Set("Authorization", tc.authorization)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Fatalf("status=%d want=%d", rec.Code, tc.want)
			}
			if tc.want == 403 && p.calls != 0 {
				t.Fatal("rejected write reached provider")
			}
		})
	}
}
func TestRemovedMembershipIsRejectedOnNextRequest(t *testing.T) {
	p := &testProvider{}
	router := gin.New()
	router.Use(NewMiddleware(p, nil).RequireAuth())
	router.GET("/profile", func(c *gin.Context) { c.Status(200) })
	for _, want := range []int{200, 401} {
		req := httptest.NewRequest("GET", "/profile", nil)
		req.Header.Set("Cookie", "better-auth.session_token=opaque")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Fatalf("status=%d want=%d", rec.Code, want)
		}
		p.revoked = true
	}
	if p.calls != 2 {
		t.Fatal("session result was cached")
	}
}
func TestRoleNameDoesNotGrantPermissions(t *testing.T) {
	i := &Identity{Roles: []Role{RoleAdmin}}
	if hasPermission(i, "org", "manage") {
		t.Fatal("client role granted access")
	}
	i.Permissions = []Permission{PermOrgManage}
	if !hasPermission(i, "org", "manage") {
		t.Fatal("explicit permission denied")
	}
}
