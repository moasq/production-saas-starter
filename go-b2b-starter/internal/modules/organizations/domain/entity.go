package domain

import "time"

// Organization represents an organization (tenant) in the system
type Organization struct {
	ID                   int32     `json:"id"`
	Slug                 string    `json:"slug"`
	Name                 string    `json:"name"`
	Status               string    `json:"status"`
	StytchOrgID          string    `json:"stytch_org_id"`
	StytchConnectionID   string    `json:"stytch_connection_id"`
	StytchConnectionName string    `json:"stytch_connection_name"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// Account represents a user account within an organization
type Account struct {
	ID                  int32      `json:"id"`
	OrganizationID      int32      `json:"organization_id"`
	Email               string     `json:"email"`
	FullName            string     `json:"full_name"`
	StytchMemberID      string     `json:"stytch_member_id"`
	StytchRoleID        string     `json:"stytch_role_id"`
	StytchRoleSlug      string     `json:"stytch_role_slug"`
	StytchEmailVerified bool       `json:"stytch_email_verified"`
	Role                string     `json:"role"`
	Status              string     `json:"status"`
	LastLoginAt         *time.Time `json:"last_login_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// OrganizationContext provides context for operations within an organization
type OrganizationContext struct {
	OrganizationID int32  `json:"organization_id"`
	AccountID      int32  `json:"account_id"`
	AccountRole    string `json:"account_role"`
}

// Implements auth.OrganizationEntity interface.
func (o *Organization) GetID() int32 {
	return o.ID
}

// Validate validates the organization entity
func (o *Organization) Validate() error {
	if o.Name == "" {
		return ErrOrganizationNameRequired
	}
	if o.Slug == "" {
		return ErrOrganizationSlugRequired
	}
	if len(o.Slug) < 3 {
		return ErrOrganizationSlugTooShort
	}
	return nil
}

// Implements auth.AccountEntity interface.
func (a *Account) GetID() int32 {
	return a.ID
}

// Validate validates the account entity
func (a *Account) Validate() error {
	if a.Email == "" {
		return ErrAccountEmailRequired
	}
	if a.FullName == "" {
		return ErrAccountFullNameRequired
	}
	if a.OrganizationID == 0 {
		return ErrAccountOrganizationRequired
	}
	return nil
}

// IsOwner checks if the account has admin role (legacy function name, kept for compatibility)
func (a *Account) IsOwner() bool {
	return a.Role == "admin"
}

// IsAdmin checks if the account has admin role
func (a *Account) IsAdmin() bool {
	return a.Role == "admin"
}

// CanManageAccounts checks if the account can manage other accounts
func (a *Account) CanManageAccounts() bool {
	return a.IsAdmin()
}
