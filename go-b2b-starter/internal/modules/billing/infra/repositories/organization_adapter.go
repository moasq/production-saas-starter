package repositories

import (
	"context"
	"fmt"

	"github.com/moasq/go-b2b-starter/internal/db/adapters"
	"github.com/moasq/go-b2b-starter/internal/modules/billing/domain"
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
