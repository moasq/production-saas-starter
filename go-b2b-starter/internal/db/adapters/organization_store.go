package adapters

import (
	"context"

	db "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
	"github.com/jackc/pgx/v5/pgtype"
)

// OrganizationStore provides database operations for organizations
type OrganizationStore interface {
	CreateOrganization(ctx context.Context, arg db.CreateOrganizationParams) (db.OrganizationsOrganization, error)
	GetOrganizationByID(ctx context.Context, id int32) (db.OrganizationsOrganization, error)
	GetOrganizationBySlug(ctx context.Context, slug string) (db.OrganizationsOrganization, error)
	GetOrganizationByStytchID(ctx context.Context, stytchOrgID pgtype.Text) (db.OrganizationsOrganization, error)
	GetOrganizationByUserEmail(ctx context.Context, email string) (db.OrganizationsOrganization, error)
	UpdateOrganization(ctx context.Context, arg db.UpdateOrganizationParams) (db.OrganizationsOrganization, error)
	UpdateOrganizationStytchInfo(ctx context.Context, arg db.UpdateOrganizationStytchInfoParams) (db.OrganizationsOrganization, error)
	ListOrganizations(ctx context.Context, arg db.ListOrganizationsParams) ([]db.OrganizationsOrganization, error)
	DeleteOrganization(ctx context.Context, id int32) error
	GetOrganizationStats(ctx context.Context, id int32) (db.GetOrganizationStatsRow, error)
}

// AccountStore provides database operations for accounts
type AccountStore interface {
	CreateAccount(ctx context.Context, arg db.CreateAccountParams) (db.OrganizationsAccount, error)
	GetAccountByID(ctx context.Context, arg db.GetAccountByIDParams) (db.OrganizationsAccount, error)
	GetAccountByEmail(ctx context.Context, arg db.GetAccountByEmailParams) (db.OrganizationsAccount, error)
	ListAccountsByOrganization(ctx context.Context, organizationID int32) ([]db.OrganizationsAccount, error)
	UpdateAccount(ctx context.Context, arg db.UpdateAccountParams) (db.OrganizationsAccount, error)
	UpdateAccountStytchInfo(ctx context.Context, arg db.UpdateAccountStytchInfoParams) (db.OrganizationsAccount, error)
	UpdateAccountLastLogin(ctx context.Context, arg db.UpdateAccountLastLoginParams) (db.OrganizationsAccount, error)
	DeleteAccount(ctx context.Context, arg db.DeleteAccountParams) error
	GetAccountOrganization(ctx context.Context, id int32) (db.OrganizationsOrganization, error)
	CheckAccountPermission(ctx context.Context, arg db.CheckAccountPermissionParams) (db.CheckAccountPermissionRow, error)
	GetAccountStats(ctx context.Context, id int32) (db.GetAccountStatsRow, error)
}
