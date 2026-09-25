package repositories

import (
	"context"
	sqlc "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
	"github.com/moasq/go-b2b-starter/internal/db/tenant"
	"github.com/moasq/go-b2b-starter/internal/modules/billing/domain"
)

type organizationAdapter struct{ store *tenant.Store }

func NewOrganizationAdapter(store *tenant.Store) domain.OrganizationAdapter {
	return &organizationAdapter{store}
}
func (a *organizationAdapter) GetExternalCustomerID(ctx context.Context, id int32) (string, error) {
	var externalID string
	err := a.store.Run(ctx, func(q *sqlc.Queries) error {
		org, err := q.GetTenantOrganization(ctx, id)
		if err != nil {
			return err
		}
		externalID = org.PolarCustomerExternalID
		return nil
	})
	return externalID, err
}
