package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/moasq/go-b2b-starter/internal/modules/organizations/domain"
	loggerDomain "github.com/moasq/go-b2b-starter/internal/platform/logger/domain"
	stytchcfg "github.com/moasq/go-b2b-starter/internal/platform/stytch"
	"github.com/stytchauth/stytch-go/v16/stytch/b2b/organizations"
)

type stytchOrganizationRepository struct {
	client       *stytchcfg.Client
	logger       loggerDomain.Logger
	localOrgRepo domain.OrganizationRepository
}

// NewStytchOrganizationRepository creates a Stytch-backed organization repository.
func NewStytchOrganizationRepository(
	client *stytchcfg.Client,
	logger loggerDomain.Logger,
	localOrgRepo domain.OrganizationRepository,
) domain.AuthOrganizationRepository {
	return &stytchOrganizationRepository{
		client:       client,
		logger:       logger,
		localOrgRepo: localOrgRepo,
	}
}

func (r *stytchOrganizationRepository) CreateOrganization(ctx context.Context, req *domain.CreateAuthOrganizationRequest) (*domain.AuthOrganization, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid create organization request: %w", err)
	}

	// Generate base slug from display name (infrastructure concern)
	baseSlug := generateSlug(req.DisplayName)

	// Prepare email invites parameter
	emailInvites := "NOT_ALLOWED"
	if req.EmailInvitesAllowed {
		emailInvites = "ALL_ALLOWED"
	}

	// Retry loop for duplicate slug handling (infrastructure concern)
	const maxAttempts = 5
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Generate slug with suffix if needed
		slug := generateSlugWithSuffix(baseSlug, attempt)

		r.logger.Debug("attempting to create organization", loggerDomain.Fields{
			"display_name": req.DisplayName,
			"slug":         slug,
			"attempt":      attempt,
		})

		// Try to create in Stytch
		params := &organizations.CreateParams{
			OrganizationSlug: slug,
			OrganizationName: req.DisplayName,
			EmailInvites:     emailInvites,
		}

		resp, err := r.client.API().Organizations.Create(ctx, params)

		// Success - return immediately
		if err == nil {
			if attempt > 1 {
				r.logger.Info("created organization with retry", loggerDomain.Fields{
					"display_name": req.DisplayName,
					"final_slug":   slug,
					"attempts":     attempt,
				})
			}
			return mapToAuthOrganization(resp.Organization), nil
		}

		// Check if duplicate slug error - retry
		if stytchcfg.IsDuplicateSlugError(err) {
			r.logger.Debug("slug already exists, retrying", loggerDomain.Fields{
				"attempted_slug": slug,
				"attempt":        attempt,
				"max_attempts":   maxAttempts,
			})
			lastErr = err
			continue // Try next suffix
		}

		// Other error - fail immediately
		return nil, fmt.Errorf("stytch create organization: %w", stytchcfg.MapError(err))
	}

	// All attempts exhausted
	r.logger.Error("failed to create organization after retries", loggerDomain.Fields{
		"display_name": req.DisplayName,
		"base_slug":    baseSlug,
		"attempts":     maxAttempts,
	})
	return nil, fmt.Errorf("failed to create organization after %d attempts, slug conflicts: %w",
		maxAttempts, stytchcfg.MapError(lastErr))
}

func (r *stytchOrganizationRepository) GetOrganization(ctx context.Context, organizationID string) (*domain.AuthOrganization, error) {
	if organizationID == "" {
		return nil, domain.ErrAuthOrganizationIDRequired
	}

	resp, err := r.client.API().Organizations.Get(ctx, &organizations.GetParams{OrganizationID: organizationID})
	if err != nil {
		return nil, fmt.Errorf("stytch get organization: %w", stytchcfg.MapError(err))
	}

	return mapToAuthOrganization(resp.Organization), nil
}

func (r *stytchOrganizationRepository) DeleteOrganization(ctx context.Context, organizationID string) error {
	if organizationID == "" {
		return domain.ErrAuthOrganizationIDRequired
	}

	_, err := r.client.API().Organizations.Delete(ctx, &organizations.DeleteParams{OrganizationID: organizationID})
	if err != nil {
		return fmt.Errorf("stytch delete organization: %w", stytchcfg.MapError(err))
	}

	return nil
}

func (r *stytchOrganizationRepository) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	if email == "" {
		return false, fmt.Errorf("email cannot be empty")
	}

	r.logger.Debug("checking if email exists", loggerDomain.Fields{
		"email": email,
	})

	// Use the local organization repository to check if email exists
	// GetByUserEmail returns organization if email is found (and account is active)
	_, err := r.localOrgRepo.GetByUserEmail(ctx, email)
	if err != nil {
		// Check if it's a "not found" error
		if err.Error() == "organization not found" || err.Error() == "sql: no rows in result set" {
			r.logger.Debug("email not found", loggerDomain.Fields{
				"email": email,
			})
			return false, nil
		}
		// Other errors are real failures
		r.logger.Error("failed to check email existence", loggerDomain.Fields{
			"email": email,
			"error": err.Error(),
		})
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	r.logger.Debug("email exists", loggerDomain.Fields{
		"email": email,
	})
	return true, nil
}

func mapToAuthOrganization(src organizations.Organization) *domain.AuthOrganization {
	var createdAt, updatedAt time.Time
	if src.CreatedAt != nil {
		createdAt = src.CreatedAt.UTC()
	}
	if src.UpdatedAt != nil {
		updatedAt = src.UpdatedAt.UTC()
	}

	return &domain.AuthOrganization{
		OrganizationID: src.OrganizationID,
		Slug:           src.OrganizationSlug,
		DisplayName:    src.OrganizationName,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}
}
