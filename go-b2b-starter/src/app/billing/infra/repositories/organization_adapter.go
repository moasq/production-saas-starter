package repositories

import (
	"context"
	"fmt"

	"github.com/moasq/go-b2b-starter/app/billing/domain"
	"github.com/moasq/go-b2b-starter/pkg/db/adapters"
	"github.com/jackc/pgx/v5/pgtype"
)

type organizationAdapter struct {
	orgStore adapters.OrganizationStore
}

func NewOrganizationAdapter(orgStore adapters.OrganizationStore) domain.OrganizationAdapter {
	return &organizationAdapter{
		orgStore: orgStore,
	}
}

func (a *organizationAdapter) GetStytchOrgID(ctx context.Context, organizationID int32) (string, error) {
	org, err := a.orgStore.GetOrganizationByID(ctx, organizationID)
	if err != nil {
		return "", fmt.Errorf("failed to get organization: %w", err)
	}

	if !org.StytchOrgID.Valid || org.StytchOrgID.String == "" {
		return "", fmt.Errorf("organization has no Stytch org ID")
	}

	return org.StytchOrgID.String, nil
}

func (a *organizationAdapter) GetOrganizationIDByStytchOrgID(ctx context.Context, stytchOrgID string) (int32, error) {
	stytchOrgIDText := pgtype.Text{
		String: stytchOrgID,
		Valid:  true,
	}

	org, err := a.orgStore.GetOrganizationByStytchID(ctx, stytchOrgIDText)
	if err != nil {
		return 0, fmt.Errorf("failed to get organization by Stytch org ID: %w", err)
	}

	return org.ID, nil
}
