package organizations

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgtype"
	sqlc "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
	"github.com/moasq/go-b2b-starter/internal/db/tenant"
	"github.com/moasq/go-b2b-starter/internal/modules/auth"
)

type Resolver struct{ store *tenant.Store }

func NewResolver(store *tenant.Store) auth.TenantResolver { return &Resolver{store} }
func (r *Resolver) Resolve(ctx context.Context, i *auth.Identity) (context.Context, *auth.RequestContext, error) {
	ctx = tenant.WithScope(ctx, i.OrganizationID)
	result := &auth.RequestContext{Identity: i, ProviderOrgID: i.OrganizationID}
	err := r.store.Run(ctx, func(q *sqlc.Queries) error {
		org, err := q.SyncOrganization(ctx, sqlc.SyncOrganizationParams{AuthOrgID: i.OrganizationID, Name: i.Organization.Name, Slug: i.Organization.Slug})
		if err != nil {
			return err
		}
		if org.Status != "active" {
			return errors.New("workspace suspended")
		}
		account, err := q.SyncAccount(ctx, sqlc.SyncAccountParams{OrganizationID: org.ID, Email: i.Email, FullName: i.User.Name, Role: string(i.Roles[0]), AuthUserID: pgtype.Text{String: i.UserID, Valid: true}, AuthMemberID: pgtype.Text{String: i.MemberID, Valid: true}})
		if err != nil {
			return err
		}
		result.OrganizationID = org.ID
		result.AccountID = account.ID
		result.CreatedAt = account.CreatedAt.Time
		result.UpdatedAt = account.UpdatedAt.Time
		return nil
	})
	return ctx, result, err
}
