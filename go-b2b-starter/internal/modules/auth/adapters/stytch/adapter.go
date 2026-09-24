// Package stytch implements the application's authentication boundary.
package stytch

import (
	"context"
	"fmt"
	"github.com/moasq/go-b2b-starter/internal/modules/auth"
	"github.com/moasq/go-b2b-starter/internal/platform/cache"
	"github.com/moasq/go-b2b-starter/internal/platform/logger"
	"github.com/stytchauth/stytch-go/v18/stytch/b2b/b2bstytchapi"
	"github.com/stytchauth/stytch-go/v18/stytch/b2b/sessions"
	"net/http"
	"time"
)

type StytchAuthAdapter struct {
	client        *b2bstytchapi.API
	policyService *RBACPolicyService
	cfg           *Config
}

func NewStytchAuthAdapter(cfg *Config, metadataCache *cache.Cache, log logger.Logger) (*StytchAuthAdapter, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	client, err := b2bstytchapi.NewClient(cfg.ProjectID, cfg.Secret, b2bstytchapi.WithSkipJWKSInitialization(), b2bstytchapi.WithBaseURI(cfg.BaseURL), b2bstytchapi.WithHTTPClient(&http.Client{Timeout: cfg.APITimeout}))
	if err != nil {
		return nil, err
	}
	return &StytchAuthAdapter{client: client, policyService: NewRBACPolicyService(client, metadataCache, log), cfg: cfg}, nil
}

// VerifyToken asks Stytch to authenticate the session. The provider is the source
// of truth for membership and permissions; an unavailable provider fails closed.
func (a *StytchAuthAdapter) VerifyToken(ctx context.Context, token string) (*auth.Identity, error) {
	if token == "" {
		return nil, auth.ErrInvalidToken
	}
	ctx, cancel := context.WithTimeout(ctx, a.cfg.APITimeout)
	defer cancel()
	result, err := a.client.Sessions.Authenticate(ctx, &sessions.AuthenticateParams{SessionJWT: token})
	if err != nil {
		return nil, auth.ErrInvalidToken
	}
	member, session := result.Member, result.MemberSession
	if member.MemberID == "" || member.MemberID != session.MemberID || member.OrganizationID != session.OrganizationID || session.OrganizationID == "" {
		return nil, auth.ErrInvalidToken
	}
	if !member.EmailAddressVerified {
		return nil, auth.ErrEmailNotVerified
	}
	if session.ExpiresAt == nil || !session.ExpiresAt.After(time.Now()) {
		return nil, auth.ErrTokenExpired
	}
	identity := &auth.Identity{UserID: member.MemberID, OrganizationID: session.OrganizationID, Email: member.EmailAddress, EmailVerified: true, ExpiresAt: *session.ExpiresAt}
	seen := map[auth.Permission]bool{}
	for _, role := range session.Roles {
		identity.Roles = append(identity.Roles, auth.NormalizeRole(role))
		permissions, err := a.policyService.GetRolePermissions(ctx, role)
		if err != nil {
			return nil, fmt.Errorf("resolve permissions: %w", err)
		}
		for _, permission := range permissions {
			if !seen[permission] {
				seen[permission] = true
				identity.Permissions = append(identity.Permissions, permission)
			}
		}
	}
	return identity, nil
}
