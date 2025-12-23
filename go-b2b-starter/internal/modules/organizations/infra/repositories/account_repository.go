package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/moasq/go-b2b-starter/internal/db/helpers"
	sqlc "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
	"github.com/moasq/go-b2b-starter/internal/modules/organizations/domain"
)

// accountRepository implements domain.AccountRepository using SQLC internally.
// SQLC types are never exposed outside this package.
type accountRepository struct {
	store sqlc.Store
}

// NewAccountRepository creates a new AccountRepository implementation.
func NewAccountRepository(store sqlc.Store) domain.AccountRepository {
	return &accountRepository{store: store}
}

func (r *accountRepository) Create(ctx context.Context, account *domain.Account) (*domain.Account, error) {
	params := sqlc.CreateAccountParams{
		OrganizationID:      account.OrganizationID,
		Email:               account.Email,
		FullName:            account.FullName,
		StytchMemberID:      helpers.ToPgText(account.StytchMemberID),
		StytchRoleID:        helpers.ToPgText(account.StytchRoleID),
		StytchRoleSlug:      helpers.ToPgText(account.StytchRoleSlug),
		StytchEmailVerified: account.StytchEmailVerified,
		Role:                account.Role,
		Status:              account.Status,
	}

	result, err := r.store.CreateAccount(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return r.mapToDomain(&result), nil
}

func (r *accountRepository) GetByID(ctx context.Context, orgID, accountID int32) (*domain.Account, error) {
	params := sqlc.GetAccountByIDParams{
		ID:             accountID,
		OrganizationID: orgID,
	}

	result, err := r.store.GetAccountByID(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to get account by ID: %w", err)
	}

	return r.mapToDomain(&result), nil
}

func (r *accountRepository) GetByEmail(ctx context.Context, orgID int32, email string) (*domain.Account, error) {
	params := sqlc.GetAccountByEmailParams{
		Email:          email,
		OrganizationID: orgID,
	}

	result, err := r.store.GetAccountByEmail(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to get account by email: %w", err)
	}

	return r.mapToDomain(&result), nil
}

func (r *accountRepository) ListByOrganization(ctx context.Context, orgID int32) ([]*domain.Account, error) {
	results, err := r.store.ListAccountsByOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts by organization: %w", err)
	}

	accounts := make([]*domain.Account, len(results))
	for i, result := range results {
		accounts[i] = r.mapToDomain(&result)
	}

	return accounts, nil
}

func (r *accountRepository) Update(ctx context.Context, account *domain.Account) (*domain.Account, error) {
	params := sqlc.UpdateAccountParams{
		ID:                  account.ID,
		OrganizationID:      account.OrganizationID,
		FullName:            account.FullName,
		StytchRoleID:        helpers.ToPgText(account.StytchRoleID),
		StytchRoleSlug:      helpers.ToPgText(account.StytchRoleSlug),
		StytchEmailVerified: account.StytchEmailVerified,
		Role:                account.Role,
		Status:              account.Status,
	}

	result, err := r.store.UpdateAccount(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	return r.mapToDomain(&result), nil
}

func (r *accountRepository) UpdateStytchInfo(ctx context.Context, orgID, accountID int32, stytchMemberID, stytchRoleID, stytchRoleSlug string, stytchEmailVerified bool) (*domain.Account, error) {
	params := sqlc.UpdateAccountStytchInfoParams{
		ID:                  accountID,
		OrganizationID:      orgID,
		StytchMemberID:      helpers.ToPgText(stytchMemberID),
		StytchRoleID:        helpers.ToPgText(stytchRoleID),
		StytchRoleSlug:      helpers.ToPgText(stytchRoleSlug),
		StytchEmailVerified: stytchEmailVerified,
	}

	result, err := r.store.UpdateAccountStytchInfo(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to update account Stytch info: %w", err)
	}

	return r.mapToDomain(&result), nil
}

func (r *accountRepository) UpdateLastLogin(ctx context.Context, orgID, accountID int32) (*domain.Account, error) {
	params := sqlc.UpdateAccountLastLoginParams{
		ID:             accountID,
		OrganizationID: orgID,
	}

	result, err := r.store.UpdateAccountLastLogin(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to update account last login: %w", err)
	}

	return r.mapToDomain(&result), nil
}

func (r *accountRepository) Delete(ctx context.Context, orgID, accountID int32) error {
	params := sqlc.DeleteAccountParams{
		ID:             accountID,
		OrganizationID: orgID,
	}

	err := r.store.DeleteAccount(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrAccountNotFound
		}
		return fmt.Errorf("failed to delete account: %w", err)
	}

	return nil
}

func (r *accountRepository) GetOrganization(ctx context.Context, accountID int32) (*domain.Organization, error) {
	result, err := r.store.GetAccountOrganization(ctx, accountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get account organization: %w", err)
	}

	return &domain.Organization{
		ID:        result.ID,
		Slug:      result.Slug,
		Name:      result.Name,
		Status:    result.Status,
		CreatedAt: result.CreatedAt.Time,
		UpdatedAt: result.UpdatedAt.Time,
	}, nil
}

func (r *accountRepository) CheckPermission(ctx context.Context, orgID, accountID int32) (*domain.AccountPermission, error) {
	params := sqlc.CheckAccountPermissionParams{
		ID:             accountID,
		OrganizationID: orgID,
	}

	result, err := r.store.CheckAccountPermission(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to check account permission: %w", err)
	}

	return &domain.AccountPermission{
		AccountID: result.ID,
		Role:      result.Role,
		Status:    result.Status,
		OrgStatus: result.OrgStatus,
	}, nil
}

func (r *accountRepository) GetStats(ctx context.Context, accountID int32) (*domain.AccountStats, error) {
	result, err := r.store.GetAccountStats(ctx, accountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to get account stats: %w", err)
	}

	account := &domain.Account{
		ID:                  result.ID,
		OrganizationID:      result.OrganizationID,
		Email:               result.Email,
		FullName:            result.FullName,
		StytchMemberID:      helpers.FromPgText(result.StytchMemberID),
		StytchRoleID:        helpers.FromPgText(result.StytchRoleID),
		StytchRoleSlug:      helpers.FromPgText(result.StytchRoleSlug),
		StytchEmailVerified: result.StytchEmailVerified,
		Role:                result.Role,
		Status:              result.Status,
		CreatedAt:           result.CreatedAt.Time,
		UpdatedAt:           result.UpdatedAt.Time,
	}

	if result.LastLoginAt.Valid {
		account.LastLoginAt = &result.LastLoginAt.Time
	}

	stats := &domain.AccountStats{
		Account:          account,
		OrganizationName: result.OrganizationName,
		OrganizationSlug: result.OrganizationSlug,
	}

	return stats, nil
}

// mapToDomain converts SQLC account type to domain type.
// This is the translation boundary - SQLC types never escape this function.
func (r *accountRepository) mapToDomain(sqlcAccount *sqlc.OrganizationsAccount) *domain.Account {
	account := &domain.Account{
		ID:                  sqlcAccount.ID,
		OrganizationID:      sqlcAccount.OrganizationID,
		Email:               sqlcAccount.Email,
		FullName:            sqlcAccount.FullName,
		StytchMemberID:      helpers.FromPgText(sqlcAccount.StytchMemberID),
		StytchRoleID:        helpers.FromPgText(sqlcAccount.StytchRoleID),
		StytchRoleSlug:      helpers.FromPgText(sqlcAccount.StytchRoleSlug),
		StytchEmailVerified: sqlcAccount.StytchEmailVerified,
		Role:                sqlcAccount.Role,
		Status:              sqlcAccount.Status,
		CreatedAt:           sqlcAccount.CreatedAt.Time,
		UpdatedAt:           sqlcAccount.UpdatedAt.Time,
	}

	// Handle nullable LastLoginAt
	if sqlcAccount.LastLoginAt.Valid {
		account.LastLoginAt = &sqlcAccount.LastLoginAt.Time
	}

	return account
}
