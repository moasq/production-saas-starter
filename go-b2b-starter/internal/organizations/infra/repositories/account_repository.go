package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/moasq/go-b2b-starter/internal/organizations/domain"
	"github.com/moasq/go-b2b-starter/internal/db/postgres"
	sqlc "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
	"github.com/moasq/go-b2b-starter/internal/db/adapters"
)

type accountRepository struct {
	accountStore adapters.AccountStore
	orgStore     adapters.OrganizationStore
}

func NewAccountRepository(accountStore adapters.AccountStore, orgStore adapters.OrganizationStore) domain.AccountRepository {
	return &accountRepository{
		accountStore: accountStore,
		orgStore:     orgStore,
	}
}

func (r *accountRepository) Create(ctx context.Context, account *domain.Account) (*domain.Account, error) {
	params := sqlc.CreateAccountParams{
		OrganizationID:      account.OrganizationID,
		Email:               account.Email,
		FullName:            account.FullName,
		StytchMemberID:      postgres.PgText(&account.StytchMemberID),
		StytchRoleID:        postgres.PgText(&account.StytchRoleID),
		StytchRoleSlug:      postgres.PgText(&account.StytchRoleSlug),
		StytchEmailVerified: account.StytchEmailVerified,
		Role:                account.Role,
		Status:              account.Status,
	}

	result, err := r.accountStore.CreateAccount(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return r.mapToDomainAccount(&result), nil
}

func (r *accountRepository) GetByID(ctx context.Context, orgID, accountID int32) (*domain.Account, error) {
	params := sqlc.GetAccountByIDParams{
		ID:             accountID,
		OrganizationID: orgID,
	}

	result, err := r.accountStore.GetAccountByID(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to get account by ID: %w", err)
	}

	return r.mapToDomainAccount(&result), nil
}

func (r *accountRepository) GetByEmail(ctx context.Context, orgID int32, email string) (*domain.Account, error) {
	params := sqlc.GetAccountByEmailParams{
		Email:          email,
		OrganizationID: orgID,
	}

	result, err := r.accountStore.GetAccountByEmail(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to get account by email: %w", err)
	}

	return r.mapToDomainAccount(&result), nil
}

func (r *accountRepository) ListByOrganization(ctx context.Context, orgID int32) ([]*domain.Account, error) {
	results, err := r.accountStore.ListAccountsByOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts by organization: %w", err)
	}

	accounts := make([]*domain.Account, len(results))
	for i, result := range results {
		accounts[i] = r.mapToDomainAccount(&result)
	}

	return accounts, nil
}

func (r *accountRepository) Update(ctx context.Context, account *domain.Account) (*domain.Account, error) {
	params := sqlc.UpdateAccountParams{
		ID:                  account.ID,
		OrganizationID:      account.OrganizationID,
		FullName:            account.FullName,
		StytchRoleID:        postgres.PgText(&account.StytchRoleID),
		StytchRoleSlug:      postgres.PgText(&account.StytchRoleSlug),
		StytchEmailVerified: account.StytchEmailVerified,
		Role:                account.Role,
		Status:              account.Status,
	}

	result, err := r.accountStore.UpdateAccount(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	return r.mapToDomainAccount(&result), nil
}

func (r *accountRepository) UpdateStytchInfo(ctx context.Context, orgID, accountID int32, stytchMemberID, stytchRoleID, stytchRoleSlug string, stytchEmailVerified bool) (*domain.Account, error) {
	params := sqlc.UpdateAccountStytchInfoParams{
		ID:                  accountID,
		OrganizationID:      orgID,
		StytchMemberID:      postgres.PgText(&stytchMemberID),
		StytchRoleID:        postgres.PgText(&stytchRoleID),
		StytchRoleSlug:      postgres.PgText(&stytchRoleSlug),
		StytchEmailVerified: stytchEmailVerified,
	}

	result, err := r.accountStore.UpdateAccountStytchInfo(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to update account Stytch info: %w", err)
	}

	return r.mapToDomainAccount(&result), nil
}

func (r *accountRepository) UpdateLastLogin(ctx context.Context, orgID, accountID int32) (*domain.Account, error) {
	params := sqlc.UpdateAccountLastLoginParams{
		ID:             accountID,
		OrganizationID: orgID,
	}

	result, err := r.accountStore.UpdateAccountLastLogin(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to update account last login: %w", err)
	}

	return r.mapToDomainAccount(&result), nil
}

func (r *accountRepository) Delete(ctx context.Context, orgID, accountID int32) error {
	params := sqlc.DeleteAccountParams{
		ID:             accountID,
		OrganizationID: orgID,
	}

	err := r.accountStore.DeleteAccount(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrAccountNotFound
		}
		return fmt.Errorf("failed to delete account: %w", err)
	}

	return nil
}

func (r *accountRepository) GetOrganization(ctx context.Context, accountID int32) (*domain.Organization, error) {
	result, err := r.accountStore.GetAccountOrganization(ctx, accountID)
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

	result, err := r.accountStore.CheckAccountPermission(ctx, params)
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
	result, err := r.accountStore.GetAccountStats(ctx, accountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to get account stats: %w", err)
	}

	// Map the result to domain stats
	account := &domain.Account{
		ID:                  result.ID,
		OrganizationID:      result.OrganizationID,
		Email:               result.Email,
		FullName:            result.FullName,
		StytchMemberID:      postgres.StringFromPgText(result.StytchMemberID),
		StytchRoleID:        postgres.StringFromPgText(result.StytchRoleID),
		StytchRoleSlug:      postgres.StringFromPgText(result.StytchRoleSlug),
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

// mapToDomainAccount maps SQLC account to domain account
func (r *accountRepository) mapToDomainAccount(sqlcAccount *sqlc.OrganizationsAccount) *domain.Account {
	account := &domain.Account{
		ID:                  sqlcAccount.ID,
		OrganizationID:      sqlcAccount.OrganizationID,
		Email:               sqlcAccount.Email,
		FullName:            sqlcAccount.FullName,
		StytchMemberID:      postgres.StringFromPgText(sqlcAccount.StytchMemberID),
		StytchRoleID:        postgres.StringFromPgText(sqlcAccount.StytchRoleID),
		StytchRoleSlug:      postgres.StringFromPgText(sqlcAccount.StytchRoleSlug),
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
