package betterauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSessionUsesLiveCookieBridgeAndRejectsInvalidIdentity(t *testing.T) {
	secret := strings.Repeat("s", 32)
	calls := 0
	valid := map[string]any{"user_id": "user-1", "member_id": "member-1", "email": "one@example.com", "email_verified": true, "organization_id": "org-1", "organization": map[string]any{"id": "org-1", "name": "One", "slug": "one"}, "roles": []string{"admin"}, "permissions": []string{"org:view", "org:manage"}, "expires_at": time.Now().Add(time.Hour)}
	status := 200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Path != "/internal/auth/session" || r.Method != "POST" || r.Header.Get("X-Internal-Auth-Secret") != secret || r.Header.Get("Cookie") != "better-auth.session_token=opaque" {
			t.Error("bridge credentials or route changed")
		}
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(valid)
	}))
	defer server.Close()
	c, err := New(server.URL, secret)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.VerifySession(context.Background(), "better-auth.session_token=opaque"); err != nil {
		t.Fatal(err)
	}
	status = 401
	if _, err := c.VerifySession(context.Background(), "better-auth.session_token=opaque"); err == nil {
		t.Fatal("revoked session accepted")
	}
	status = 200
	valid["expires_at"] = time.Now().Add(-time.Second)
	if _, err := c.VerifySession(context.Background(), "better-auth.session_token=opaque"); err == nil {
		t.Fatal("expired session accepted")
	}
	valid["expires_at"] = time.Now().Add(time.Hour)
	valid["organization_id"] = "org-2"
	if _, err := c.VerifySession(context.Background(), "better-auth.session_token=opaque"); err == nil {
		t.Fatal("mismatched tenant accepted")
	}
	valid["organization_id"] = "org-1"
	valid["email_verified"] = false
	if _, err := c.VerifySession(context.Background(), "better-auth.session_token=opaque"); err == nil {
		t.Fatal("unverified email accepted")
	}
	if calls != 5 {
		t.Fatalf("expected every request to recheck session, got %d", calls)
	}
}
func TestBridgeFailsClosedAndDoesNotFollowRedirects(t *testing.T) {
	reached := false
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reached = true }))
	defer destination.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, destination.URL, 307) }))
	defer server.Close()
	c, _ := New(server.URL, strings.Repeat("s", 32))
	if _, err := c.VerifySession(context.Background(), "cookie"); err == nil {
		t.Fatal("redirect accepted")
	}
	if reached {
		t.Fatal("shared secret forwarded to redirect")
	}
	if _, err := New(server.URL, ""); err == nil {
		t.Fatal("unconfigured bridge accepted")
	}
	server.Close()
	if _, err := c.VerifySession(context.Background(), "cookie"); err == nil {
		t.Fatal("unavailable auth accepted")
	}
}
