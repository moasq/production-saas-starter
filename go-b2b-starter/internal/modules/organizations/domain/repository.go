package domain

import "context"

// OrganizationRepository defines the interface for organization data operations
type OrganizationRepository interface {
	Create(ctx context.Context, org *Organization) (*Organization, error)
	GetByID(ctx context.Context, id int32) (*Organization, error)
	GetBySlug(ctx context.Context, slug string) (*Organization, error)
	GetByStytchID(ctx context.Context, stytchOrgID string) (*Organization, error)
	GetByUserEmail(ctx context.Context, email string) (*Organization, error)
	Update(ctx context.Context, org *Organization) (*Organization, error)
	UpdateStytchInfo(ctx context.Context, id int32, stytchOrgID, stytchConnectionID, stytchConnectionName string) (*Organization, error)
	Delete(ctx context.Context, id int32) error
	List(ctx context.Context, limit, offset int32) ([]*Organization, error)
	GetStats(ctx context.Context, id int32) (*OrganizationStats, error)
}

// AccountRepository defines the interface for account data operations
type AccountRepository interface {
	Create(ctx context.Context, account *Account) (*Account, error)
	GetByID(ctx context.Context, orgID, accountID int32) (*Account, error)
	GetByEmail(ctx context.Context, orgID int32, email string) (*Account, error)
	ListByOrganization(ctx context.Context, orgID int32) ([]*Account, error)
	Update(ctx context.Context, account *Account) (*Account, error)
	UpdateStytchInfo(ctx context.Context, orgID, accountID int32, stytchMemberID, stytchRoleID, stytchRoleSlug string, stytchEmailVerified bool) (*Account, error)
	UpdateLastLogin(ctx context.Context, orgID, accountID int32) (*Account, error)
	Delete(ctx context.Context, orgID, accountID int32) error
	GetOrganization(ctx context.Context, accountID int32) (*Organization, error)
	CheckPermission(ctx context.Context, orgID, accountID int32) (*AccountPermission, error)
	GetStats(ctx context.Context, accountID int32) (*AccountStats, error)
}

// OrganizationStats represents organization statistics
type OrganizationStats struct {
	Organization       *Organization `json:"organization"`
	AccountCount       int64         `json:"account_count"`
	ActiveAccountCount int64         `json:"active_account_count"`
}

// AccountStats represents account statistics with organization info
type AccountStats struct {
	Account          *Account `json:"account"`
	OrganizationName string   `json:"organization_name"`
	OrganizationSlug string   `json:"organization_slug"`
}

// AccountPermission represents account permission check result
type AccountPermission struct {
	AccountID int32  `json:"account_id"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	OrgStatus string `json:"org_status"`
}
