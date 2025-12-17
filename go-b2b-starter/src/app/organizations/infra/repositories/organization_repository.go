package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/moasq/go-b2b-starter/app/organizations/domain"
	"github.com/moasq/go-b2b-starter/pkg/db/postgres"
	sqlc "github.com/moasq/go-b2b-starter/pkg/db/postgres/sqlc/gen"
	"github.com/moasq/go-b2b-starter/pkg/db/adapters"
)

type organizationRepository struct {
	orgStore adapters.OrganizationStore
}

func NewOrganizationRepository(orgStore adapters.OrganizationStore) domain.OrganizationRepository {
	return &organizationRepository{
		orgStore: orgStore,
	}
}

func (r *organizationRepository) Create(ctx context.Context, org *domain.Organization) (*domain.Organization, error) {
	params := sqlc.CreateOrganizationParams{
		Slug:   org.Slug,
		Name:   org.Name,
		Status: org.Status,
	}

	result, err := r.orgStore.CreateOrganization(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	return r.mapToDomainOrganization(&result), nil
}

func (r *organizationRepository) GetByID(ctx context.Context, id int32) (*domain.Organization, error) {
	result, err := r.orgStore.GetOrganizationByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get organization by ID: %w", err)
	}

	return r.mapToDomainOrganization(&result), nil
}

func (r *organizationRepository) GetBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	result, err := r.orgStore.GetOrganizationBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get organization by slug: %w", err)
	}

	return r.mapToDomainOrganization(&result), nil
}

func (r *organizationRepository) GetByStytchID(ctx context.Context, stytchOrgID string) (*domain.Organization, error) {
	result, err := r.orgStore.GetOrganizationByStytchID(ctx, postgres.PgText(&stytchOrgID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get organization by Stytch ID: %w", err)
	}

	return r.mapToDomainOrganization(&result), nil
}

func (r *organizationRepository) GetByUserEmail(ctx context.Context, email string) (*domain.Organization, error) {
	result, err := r.orgStore.GetOrganizationByUserEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get organization by user email: %w", err)
	}

	return r.mapToDomainOrganization(&result), nil
}

func (r *organizationRepository) Update(ctx context.Context, org *domain.Organization) (*domain.Organization, error) {
	params := sqlc.UpdateOrganizationParams{
		ID:                   org.ID,
		Name:                 org.Name,
		Status:               org.Status,
		StytchOrgID:          postgres.PgText(&org.StytchOrgID),
		StytchConnectionID:   postgres.PgText(&org.StytchConnectionID),
		StytchConnectionName: postgres.PgText(&org.StytchConnectionName),
	}

	result, err := r.orgStore.UpdateOrganization(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	return r.mapToDomainOrganization(&result), nil
}

func (r *organizationRepository) UpdateStytchInfo(ctx context.Context, id int32, stytchOrgID, stytchConnectionID, stytchConnectionName string) (*domain.Organization, error) {
	params := sqlc.UpdateOrganizationStytchInfoParams{
		ID:                   id,
		StytchOrgID:          postgres.PgText(&stytchOrgID),
		StytchConnectionID:   postgres.PgText(&stytchConnectionID),
		StytchConnectionName: postgres.PgText(&stytchConnectionName),
	}

	result, err := r.orgStore.UpdateOrganizationStytchInfo(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to update organization Stytch info: %w", err)
	}

	return r.mapToDomainOrganization(&result), nil
}

func (r *organizationRepository) List(ctx context.Context, limit, offset int32) ([]*domain.Organization, error) {
	params := sqlc.ListOrganizationsParams{
		Limit:  limit,
		Offset: offset,
	}

	results, err := r.orgStore.ListOrganizations(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	organizations := make([]*domain.Organization, len(results))
	for i, result := range results {
		organizations[i] = r.mapToDomainOrganization(&result)
	}

	return organizations, nil
}

func (r *organizationRepository) Delete(ctx context.Context, id int32) error {
	if err := r.orgStore.DeleteOrganization(ctx, id); err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}
	return nil
}

func (r *organizationRepository) GetStats(ctx context.Context, id int32) (*domain.OrganizationStats, error) {
	result, err := r.orgStore.GetOrganizationStats(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get organization stats: %w", err)
	}

	// Map the result to domain stats
	org := &domain.Organization{
		ID:                   result.ID,
		Slug:                 result.Slug,
		Name:                 result.Name,
		Status:               result.Status,
		StytchOrgID:          postgres.StringFromPgText(result.StytchOrgID),
		StytchConnectionID:   postgres.StringFromPgText(result.StytchConnectionID),
		StytchConnectionName: postgres.StringFromPgText(result.StytchConnectionName),
		CreatedAt:            result.CreatedAt.Time,
		UpdatedAt:            result.UpdatedAt.Time,
	}

	stats := &domain.OrganizationStats{
		Organization:       org,
		AccountCount:       result.AccountCount,
		ActiveAccountCount: result.ActiveAccountCount,
	}

	return stats, nil
}

// mapToDomainOrganization maps SQLC OrganizationsOrganization to domain organization
func (r *organizationRepository) mapToDomainOrganization(sqlcOrg *sqlc.OrganizationsOrganization) *domain.Organization {
	org := &domain.Organization{
		ID:        sqlcOrg.ID,
		Slug:      sqlcOrg.Slug,
		Name:      sqlcOrg.Name,
		Status:    sqlcOrg.Status,
		CreatedAt: sqlcOrg.CreatedAt.Time,
		UpdatedAt: sqlcOrg.UpdatedAt.Time,
	}

	// Map Stytch fields
	if sqlcOrg.StytchOrgID.Valid {
		org.StytchOrgID = sqlcOrg.StytchOrgID.String
	}
	if sqlcOrg.StytchConnectionID.Valid {
		org.StytchConnectionID = sqlcOrg.StytchConnectionID.String
	}
	if sqlcOrg.StytchConnectionName.Valid {
		org.StytchConnectionName = sqlcOrg.StytchConnectionName.String
	}

	return org
}
