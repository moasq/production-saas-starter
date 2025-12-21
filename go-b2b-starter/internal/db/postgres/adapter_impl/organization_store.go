package adapterimpl

import (
	"context"

	sqlc "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
	"github.com/moasq/go-b2b-starter/internal/db/adapters"
	"github.com/jackc/pgx/v5/pgtype"
)

// organizationStore implements adapters.OrganizationStore
type organizationStore struct {
	store sqlc.Store
}

func NewOrganizationStore(store sqlc.Store) adapters.OrganizationStore {
	return &organizationStore{store: store}
}

func (s *organizationStore) GetOrganizationByID(ctx context.Context, id int32) (sqlc.OrganizationsOrganization, error) {
	return s.store.GetOrganizationByID(ctx, id)
}

func (s *organizationStore) GetOrganizationBySlug(ctx context.Context, slug string) (sqlc.OrganizationsOrganization, error) {
	return s.store.GetOrganizationBySlug(ctx, slug)
}

func (s *organizationStore) GetOrganizationByStytchID(ctx context.Context, stytchOrgID pgtype.Text) (sqlc.OrganizationsOrganization, error) {
	return s.store.GetOrganizationByStytchID(ctx, stytchOrgID)
}

func (s *organizationStore) GetOrganizationByUserEmail(ctx context.Context, email string) (sqlc.OrganizationsOrganization, error) {
	return s.store.GetOrganizationByUserEmail(ctx, email)
}

func (s *organizationStore) CreateOrganization(ctx context.Context, arg sqlc.CreateOrganizationParams) (sqlc.OrganizationsOrganization, error) {
	return s.store.CreateOrganization(ctx, arg)
}

func (s *organizationStore) UpdateOrganization(ctx context.Context, arg sqlc.UpdateOrganizationParams) (sqlc.OrganizationsOrganization, error) {
	return s.store.UpdateOrganization(ctx, arg)
}

func (s *organizationStore) UpdateOrganizationStytchInfo(ctx context.Context, arg sqlc.UpdateOrganizationStytchInfoParams) (sqlc.OrganizationsOrganization, error) {
	return s.store.UpdateOrganizationStytchInfo(ctx, arg)
}

func (s *organizationStore) ListOrganizations(ctx context.Context, arg sqlc.ListOrganizationsParams) ([]sqlc.OrganizationsOrganization, error) {
	return s.store.ListOrganizations(ctx, arg)
}

func (s *organizationStore) DeleteOrganization(ctx context.Context, id int32) error {
	return s.store.DeleteOrganization(ctx, id)
}

func (s *organizationStore) GetOrganizationStats(ctx context.Context, id int32) (sqlc.GetOrganizationStatsRow, error) {
	return s.store.GetOrganizationStats(ctx, id)
}

// accountStore implements adapters.AccountStore
type accountStore struct {
	store sqlc.Store
}

func NewAccountStore(store sqlc.Store) adapters.AccountStore {
	return &accountStore{store: store}
}

func (s *accountStore) CreateAccount(ctx context.Context, arg sqlc.CreateAccountParams) (sqlc.OrganizationsAccount, error) {
	return s.store.CreateAccount(ctx, arg)
}

func (s *accountStore) GetAccountByID(ctx context.Context, arg sqlc.GetAccountByIDParams) (sqlc.OrganizationsAccount, error) {
	return s.store.GetAccountByID(ctx, arg)
}

func (s *accountStore) GetAccountByEmail(ctx context.Context, arg sqlc.GetAccountByEmailParams) (sqlc.OrganizationsAccount, error) {
	return s.store.GetAccountByEmail(ctx, arg)
}

func (s *accountStore) ListAccountsByOrganization(ctx context.Context, organizationID int32) ([]sqlc.OrganizationsAccount, error) {
	return s.store.ListAccountsByOrganization(ctx, organizationID)
}

func (s *accountStore) UpdateAccount(ctx context.Context, arg sqlc.UpdateAccountParams) (sqlc.OrganizationsAccount, error) {
	return s.store.UpdateAccount(ctx, arg)
}

func (s *accountStore) UpdateAccountStytchInfo(ctx context.Context, arg sqlc.UpdateAccountStytchInfoParams) (sqlc.OrganizationsAccount, error) {
	return s.store.UpdateAccountStytchInfo(ctx, arg)
}

func (s *accountStore) UpdateAccountLastLogin(ctx context.Context, arg sqlc.UpdateAccountLastLoginParams) (sqlc.OrganizationsAccount, error) {
	return s.store.UpdateAccountLastLogin(ctx, arg)
}

func (s *accountStore) DeleteAccount(ctx context.Context, arg sqlc.DeleteAccountParams) error {
	return s.store.DeleteAccount(ctx, arg)
}

func (s *accountStore) GetAccountOrganization(ctx context.Context, id int32) (sqlc.OrganizationsOrganization, error) {
	return s.store.GetAccountOrganization(ctx, id)
}

func (s *accountStore) CheckAccountPermission(ctx context.Context, arg sqlc.CheckAccountPermissionParams) (sqlc.CheckAccountPermissionRow, error) {
	return s.store.CheckAccountPermission(ctx, arg)
}

func (s *accountStore) GetAccountStats(ctx context.Context, id int32) (sqlc.GetAccountStatsRow, error) {
	return s.store.GetAccountStats(ctx, id)
}
