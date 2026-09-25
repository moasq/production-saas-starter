package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moasq/go-b2b-starter/internal/db/postgres"
	"github.com/moasq/go-b2b-starter/internal/db/tenant"
	"go.uber.org/dig"
)

func Inject(container *dig.Container) error {
	for _, provider := range []any{postgres.LoadConfig, postgres.InitDB, tenant.New} {
		if err := container.Provide(provider); err != nil {
			return err
		}
	}
	return container.Invoke(func(pool *pgxpool.Pool) error { return postgres.ValidateRuntime(context.Background(), pool) })
}
