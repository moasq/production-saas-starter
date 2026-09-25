package betterauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/moasq/go-b2b-starter/internal/modules/auth"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	secret  string
	http    *http.Client
}

func NewFromEnvironment() (*Client, error) {
	return New(os.Getenv("AUTH_BRIDGE_URL"), os.Getenv("AUTH_INTERNAL_SECRET"))
}
func New(baseURL, secret string) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, fmt.Errorf("valid AUTH_BRIDGE_URL required")
	}
	if len(secret) < 32 {
		return nil, fmt.Errorf("AUTH_INTERNAL_SECRET must contain at least 32 characters")
	}
	return &Client{strings.TrimRight(baseURL, "/"), secret, &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}
func (c *Client) Call(ctx context.Context, cookie, operation string, payload any) (int, []byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/auth/"+operation, bytes.NewReader(body))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Auth-Secret", c.secret)
	req.Header.Set("Cookie", cookie)
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("auth bridge unavailable: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return 0, nil, err
	}
	return resp.StatusCode, data, nil
}
func (c *Client) VerifySession(ctx context.Context, cookie string) (*auth.Identity, error) {
	status, data, err := c.Call(ctx, cookie, "session", struct{}{})
	if err != nil || status != 200 {
		return nil, auth.ErrUnauthorized
	}
	var identity auth.Identity
	if json.Unmarshal(data, &identity) != nil || identity.UserID == "" || identity.MemberID == "" || identity.Email == "" || identity.OrganizationID == "" || identity.Organization.ID != identity.OrganizationID || !identity.EmailVerified || !identity.ExpiresAt.After(time.Now()) || len(identity.Roles) != 1 {
		return nil, auth.ErrUnauthorized
	}
	switch identity.Roles[0] {
	case auth.RoleAdmin, auth.RoleManager, auth.RoleMember:
	default:
		return nil, auth.ErrUnauthorized
	}
	// The bridge supplies explicit permissions; the Go adapter never synthesizes grants from a role name.
	for _, permission := range identity.Permissions {
		if permission != auth.PermOrgView && permission != auth.PermOrgManage {
			return nil, auth.ErrUnauthorized
		}
	}
	return &identity, nil
}
