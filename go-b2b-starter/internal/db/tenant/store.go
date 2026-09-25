package tenant

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	sqlc "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
)

type contextKey struct{}

var ErrMissingScope = errors.New("verified tenant scope required")

type Store struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

// WithScope is called only after live session and membership verification.
func WithScope(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextKey{}, id)
}
func (s *Store) Run(ctx context.Context, fn func(*sqlc.Queries) error) error {
	id, _ := ctx.Value(contextKey{}).(string)
	if id == "" {
		return ErrMissingScope
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.WithoutCancel(ctx))
	// true makes the setting transaction-local, including rollback and pooled reuse.
	if _, err = tx.Exec(ctx, "SELECT set_config('app.tenant_id',$1,true)", id); err != nil {
		return err
	}
	if err = fn(sqlc.New(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
