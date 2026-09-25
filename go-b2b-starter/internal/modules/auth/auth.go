package auth

import (
	"context"
	"time"
)

// AuthProvider checks the opaque cookie against the live self-hosted session store.
type AuthProvider interface {
	VerifySession(context.Context, string) (*Identity, error)
}
type Identity struct {
	UserID         string       `json:"user_id"`
	MemberID       string       `json:"member_id"`
	Email          string       `json:"email"`
	EmailVerified  bool         `json:"email_verified"`
	OrganizationID string       `json:"organization_id"`
	Roles          []Role       `json:"roles"`
	Permissions    []Permission `json:"permissions"`
	ExpiresAt      time.Time    `json:"expires_at"`
	Organization   struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"organization"`
	User struct {
		Name string `json:"name"`
	} `json:"user"`
}

func (i *Identity) HasPermission(p Permission) bool {
	for _, v := range i.Permissions {
		if v == p {
			return true
		}
	}
	return false
}

type RequestContext struct {
	Identity       *Identity
	OrganizationID int32
	AccountID      int32
	ProviderOrgID  string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// TenantResolver mirrors the already verified identity and supplies database scope.
type TenantResolver interface {
	Resolve(context.Context, *Identity) (context.Context, *RequestContext, error)
}
