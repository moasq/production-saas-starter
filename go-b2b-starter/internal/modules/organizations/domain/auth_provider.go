package domain

import (
	"context"
	"net/mail"
	"time"
)

// AuthMember represents an authenticated member from the auth provider.
type AuthMember struct {
	MemberID       string    `json:"member_id"`
	OrganizationID string    `json:"organization_id"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	Roles          []string  `json:"roles"`
	Status         string    `json:"status"`
	EmailVerified  bool      `json:"email_verified"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// AuthOrganization represents an organization (tenant) from the auth provider.
type AuthOrganization struct {
	OrganizationID string    `json:"organization_id"`
	Slug           string    `json:"slug"`
	DisplayName    string    `json:"display_name"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// AuthRole represents an RBAC role from the auth provider.
type AuthRole struct {
	RoleID      string   `json:"role_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// CreateAuthMemberRequest represents the data needed to create a member in the auth provider.
type CreateAuthMemberRequest struct {
	OrganizationID string   `json:"organization_id"`
	Email          string   `json:"email"`
	Name           string   `json:"name"`
	Roles          []string `json:"roles"`
	SendInvite     bool     `json:"send_invite"`
	Password       string   `json:"password"`
}

// UpdateAuthMemberRequest represents member profile updates in the auth provider.
type UpdateAuthMemberRequest struct {
	OrganizationID string         `json:"organization_id"`
	MemberID       string         `json:"member_id"`
	Name           *string        `json:"name,omitempty"`
	Roles          []string       `json:"roles,omitempty"`
	TrustedMeta    map[string]any `json:"trusted_metadata,omitempty"`
	UntrustedMeta  map[string]any `json:"untrusted_metadata,omitempty"`
}

// CreateAuthOrganizationRequest represents the data needed to create an organization in the auth provider.
type CreateAuthOrganizationRequest struct {
	DisplayName         string `json:"display_name"`
	EmailInvitesAllowed bool   `json:"email_invites_allowed"`
}

// AssignAuthRolesRequest represents assigning roles to a member in the auth provider.
type AssignAuthRolesRequest struct {
	OrganizationID string   `json:"organization_id"`
	MemberID       string   `json:"member_id"`
	Roles          []string `json:"roles"`
}

// RemoveAuthMembersRequest represents removing members from an organization in the auth provider.
type RemoveAuthMembersRequest struct {
	OrganizationID string   `json:"organization_id"`
	MemberIDs      []string `json:"member_ids"`
}

// SendMagicLinkRequest represents the payload required to email a login magic link.
type SendMagicLinkRequest struct {
	OrganizationID    string `json:"organization_id"`
	Email             string `json:"email"`
	LoginRedirectURL  string `json:"login_redirect_url"`
	SignupRedirectURL string `json:"signup_redirect_url"`
}

// Validate validates the CreateAuthMemberRequest.
func (r *CreateAuthMemberRequest) Validate() error {
	if r.OrganizationID == "" {
		return ErrAuthOrganizationIDRequired
	}
	if r.Email == "" {
		return ErrAuthEmailRequired
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		return ErrAuthInvalidEmail
	}
	if r.Name == "" {
		return ErrAuthNameRequired
	}
	return nil
}

// Validate ensures the SendMagicLinkRequest contains core identifiers.
func (r *SendMagicLinkRequest) Validate() error {
	if r.OrganizationID == "" {
		return ErrAuthOrganizationIDRequired
	}
	if r.Email == "" {
		return ErrAuthEmailRequired
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		return ErrAuthInvalidEmail
	}
	return nil
}

// Validate validates the UpdateAuthMemberRequest.
func (r *UpdateAuthMemberRequest) Validate() error {
	if r.OrganizationID == "" {
		return ErrAuthOrganizationIDRequired
	}
	if r.MemberID == "" {
		return ErrAuthMemberIDRequired
	}
	return nil
}

// Validate validates the CreateAuthOrganizationRequest.
func (r *CreateAuthOrganizationRequest) Validate() error {
	if r.DisplayName == "" {
		return ErrAuthOrganizationDisplayNameRequired
	}
	if len(r.DisplayName) < 2 {
		return ErrAuthOrganizationNameTooShort
	}
	return nil
}

// Validate validates the AssignAuthRolesRequest.
func (r *AssignAuthRolesRequest) Validate() error {
	if r.OrganizationID == "" {
		return ErrAuthOrganizationIDRequired
	}
	if r.MemberID == "" {
		return ErrAuthMemberIDRequired
	}
	if len(r.Roles) == 0 {
		return ErrAuthRoleIDsRequired
	}
	return nil
}

// Validate validates the RemoveAuthMembersRequest.
func (r *RemoveAuthMembersRequest) Validate() error {
	if r.OrganizationID == "" {
		return ErrAuthOrganizationIDRequired
	}
	if len(r.MemberIDs) == 0 {
		return ErrAuthMemberIDsRequired
	}
	return nil
}

// AuthOrganizationRepository defines auth provider organization operations.
type AuthOrganizationRepository interface {
	CreateOrganization(ctx context.Context, req *CreateAuthOrganizationRequest) (*AuthOrganization, error)
	GetOrganization(ctx context.Context, organizationID string) (*AuthOrganization, error)
	DeleteOrganization(ctx context.Context, organizationID string) error
	CheckEmailExists(ctx context.Context, email string) (bool, error)
}

// AuthMemberRepository defines auth provider member operations.
type AuthMemberRepository interface {
	CreateMember(ctx context.Context, req *CreateAuthMemberRequest) (*AuthMember, error)
	UpdateMember(ctx context.Context, req *UpdateAuthMemberRequest) (*AuthMember, error)
	GetMember(ctx context.Context, organizationID, memberID string) (*AuthMember, error)
	GetMemberByEmail(ctx context.Context, organizationID, email string) (*AuthMember, error)
	ListMembers(ctx context.Context, organizationID string, limit, offset int) ([]*AuthMember, error)
	RemoveMembers(ctx context.Context, req *RemoveAuthMembersRequest) error
	AssignRoles(ctx context.Context, req *AssignAuthRolesRequest) error
	SendMagicLink(ctx context.Context, req *SendMagicLinkRequest) error
}

// AuthRoleRepository defines auth provider RBAC operations.
type AuthRoleRepository interface {
	GetRoleByID(ctx context.Context, roleID string) (*AuthRole, error)
	GetRoleBySlug(ctx context.Context, slug string) (*AuthRole, error)
	ListRoles(ctx context.Context, limit, offset int) ([]*AuthRole, error)
}
