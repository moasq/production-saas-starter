package services

import (
	"context"

	"github.com/moasq/go-b2b-starter/internal/organizations/domain"
)

// OrganizationService defines the interface for organization business operations
type OrganizationService interface {
	// Organization operations
	CreateOrganization(ctx context.Context, req *CreateOrganizationRequest) (*domain.Organization, error)
	GetOrganization(ctx context.Context, orgID int32) (*domain.Organization, error)
	GetOrganizationBySlug(ctx context.Context, slug string) (*domain.Organization, error)
	GetOrganizationByStytchID(ctx context.Context, stytchOrgID string) (*domain.Organization, error)
	GetOrganizationByUserEmail(ctx context.Context, email string) (*domain.Organization, error)
	UpdateOrganization(ctx context.Context, orgID int32, req *UpdateOrganizationRequest) (*domain.Organization, error)
	ListOrganizations(ctx context.Context, req *ListOrganizationsRequest) (*ListOrganizationsResponse, error)
	GetOrganizationStats(ctx context.Context, orgID int32) (*domain.OrganizationStats, error)

	// Account operations
	CreateAccount(ctx context.Context, orgID int32, req *CreateAccountRequest) (*domain.Account, error)
	GetAccount(ctx context.Context, orgID, accountID int32) (*domain.Account, error)
	GetAccountByEmail(ctx context.Context, orgID int32, email string) (*domain.Account, error)
	ListAccounts(ctx context.Context, orgID int32) ([]*domain.Account, error)
	UpdateAccount(ctx context.Context, orgID, accountID int32, req *UpdateAccountRequest) (*domain.Account, error)
	DeleteAccount(ctx context.Context, orgID, accountID int32) error
	UpdateAccountLastLogin(ctx context.Context, orgID, accountID int32) (*domain.Account, error)

	// Utility operations
	CheckAccountPermission(ctx context.Context, orgID, accountID int32) (*domain.AccountPermission, error)
	GetAccountStats(ctx context.Context, accountID int32) (*domain.AccountStats, error)
}

// CreateOrganizationRequest represents data needed to create an organization
type CreateOrganizationRequest struct {
	Slug                 string `json:"slug" binding:"required,min=3"`
	Name                 string `json:"name" binding:"required"`
	OwnerEmail           string `json:"owner_email" binding:"required,email"`
	OwnerName            string `json:"owner_name" binding:"required"`
	StytchOrgID          string `json:"stytch_org_id"`
	StytchConnectionID   string `json:"stytch_connection_id"`
	StytchConnectionName string `json:"stytch_connection_name"`
}

// UpdateOrganizationRequest represents data needed to update an organization
type UpdateOrganizationRequest struct {
	Name                 string `json:"name" binding:"required"`
	Status               string `json:"status" binding:"required,oneof=active suspended"`
	StytchOrgID          string `json:"stytch_org_id"`
	StytchConnectionID   string `json:"stytch_connection_id"`
	StytchConnectionName string `json:"stytch_connection_name"`
}

// CreateAccountRequest represents data needed to create an account
type CreateAccountRequest struct {
	Email               string `json:"email" binding:"required,email"`
	FullName            string `json:"full_name" binding:"required"`
	Role                string `json:"role" binding:"required,oneof=admin approver member"`
	StytchMemberID      string `json:"stytch_member_id"`
	StytchRoleID        string `json:"stytch_role_id"`
	StytchRoleSlug      string `json:"stytch_role_slug"`
	StytchEmailVerified bool   `json:"stytch_email_verified"`
}

// UpdateAccountRequest represents data needed to update an account
type UpdateAccountRequest struct {
	FullName            string `json:"full_name" binding:"required"`
	Role                string `json:"role" binding:"required,oneof=admin approver member"`
	Status              string `json:"status" binding:"required,oneof=active inactive suspended"`
	StytchRoleID        string `json:"stytch_role_id"`
	StytchRoleSlug      string `json:"stytch_role_slug"`
	StytchEmailVerified *bool  `json:"stytch_email_verified"`
}

// ListOrganizationsRequest represents parameters for listing organizations
type ListOrganizationsRequest struct {
	Limit  int32 `json:"limit" binding:"min=1,max=100"`
	Offset int32 `json:"offset" binding:"min=0"`
}

// ListOrganizationsResponse represents the response for listing organizations
type ListOrganizationsResponse struct {
	Organizations []*domain.Organization `json:"organizations"`
	Total         int32                  `json:"total"`
	Limit         int32                  `json:"limit"`
	Offset        int32                  `json:"offset"`
}
