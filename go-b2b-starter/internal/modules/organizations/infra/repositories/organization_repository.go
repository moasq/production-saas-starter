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

// organizationRepository implements domain.OrganizationRepository using SQLC internally.
// SQLC types are never exposed outside this package.
type organizationRepository struct {
	store sqlc.Store
}

// NewOrganizationRepository creates a new OrganizationRepository implementation.
func NewOrganizationRepository(store sqlc.Store) domain.OrganizationRepository {
	return &organizationRepository{store: store}
}

func (r *organizationRepository) Create(ctx context.Context, org *domain.Organization) (*domain.Organization, error) {
	params := sqlc.CreateOrganizationParams{
		Slug:   org.Slug,
		Name:   org.Name,
		Status: org.Status,
	}

	result, err := r.store.CreateOrganization(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	return r.mapToDomain(&result), nil
}

func (r *organizationRepository) GetByID(ctx context.Context, id int32) (*domain.Organization, error) {
	result, err := r.store.GetOrganizationByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get organization by ID: %w", err)
	}

	return r.mapToDomain(&result), nil
}

func (r *organizationRepository) GetBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	result, err := r.store.GetOrganizationBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get organization by slug: %w", err)
	}

	return r.mapToDomain(&result), nil
}

func (r *organizationRepository) GetByStytchID(ctx context.Context, stytchOrgID string) (*domain.Organization, error) {
	result, err := r.store.GetOrganizationByStytchID(ctx, helpers.ToPgText(stytchOrgID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get organization by Stytch ID: %w", err)
	}

	return r.mapToDomain(&result), nil
}

func (r *organizationRepository) GetByUserEmail(ctx context.Context, email string) (*domain.Organization, error) {
	result, err := r.store.GetOrganizationByUserEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get organization by user email: %w", err)
	}

	return r.mapToDomain(&result), nil
}

func (r *organizationRepository) Update(ctx context.Context, org *domain.Organization) (*domain.Organization, error) {
	params := sqlc.UpdateOrganizationParams{
		ID:                   org.ID,
		Name:                 org.Name,
		Status:               org.Status,
		StytchOrgID:          helpers.ToPgText(org.StytchOrgID),
		StytchConnectionID:   helpers.ToPgText(org.StytchConnectionID),
		StytchConnectionName: helpers.ToPgText(org.StytchConnectionName),
	}

	result, err := r.store.UpdateOrganization(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	return r.mapToDomain(&result), nil
}

func (r *organizationRepository) UpdateStytchInfo(ctx context.Context, id int32, stytchOrgID, stytchConnectionID, stytchConnectionName string) (*domain.Organization, error) {
	params := sqlc.UpdateOrganizationStytchInfoParams{
		ID:                   id,
		StytchOrgID:          helpers.ToPgText(stytchOrgID),
		StytchConnectionID:   helpers.ToPgText(stytchConnectionID),
		StytchConnectionName: helpers.ToPgText(stytchConnectionName),
	}

	result, err := r.store.UpdateOrganizationStytchInfo(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to update organization Stytch info: %w", err)
	}

	return r.mapToDomain(&result), nil
}

func (r *organizationRepository) List(ctx context.Context, limit, offset int32) ([]*domain.Organization, error) {
	params := sqlc.ListOrganizationsParams{
		Limit:  limit,
		Offset: offset,
	}

	results, err := r.store.ListOrganizations(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	organizations := make([]*domain.Organization, len(results))
	for i, result := range results {
		organizations[i] = r.mapToDomain(&result)
	}

	return organizations, nil
}

func (r *organizationRepository) Delete(ctx context.Context, id int32) error {
	if err := r.store.DeleteOrganization(ctx, id); err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}
	return nil
}

func (r *organizationRepository) GetStats(ctx context.Context, id int32) (*domain.OrganizationStats, error) {
	result, err := r.store.GetOrganizationStats(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("failed to get organization stats: %w", err)
	}

	org := &domain.Organization{
		ID:                   result.ID,
		Slug:                 result.Slug,
		Name:                 result.Name,
		Status:               result.Status,
		StytchOrgID:          helpers.FromPgText(result.StytchOrgID),
		StytchConnectionID:   helpers.FromPgText(result.StytchConnectionID),
		StytchConnectionName: helpers.FromPgText(result.StytchConnectionName),
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

// mapToDomain converts SQLC organization type to domain type.
// This is the translation boundary - SQLC types never escape this function.
func (r *organizationRepository) mapToDomain(sqlcOrg *sqlc.OrganizationsOrganization) *domain.Organization {
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
