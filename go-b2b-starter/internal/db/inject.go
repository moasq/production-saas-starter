package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/dig"

	// Domain interfaces - these are the interfaces we provide
	orgDomain "github.com/moasq/go-b2b-starter/internal/modules/organizations/domain"

	// Repository implementations from module infra layers
	orgRepos "github.com/moasq/go-b2b-starter/internal/modules/organizations/infra/repositories"

	// Legacy adapters - kept temporarily for backward compatibility
	"github.com/moasq/go-b2b-starter/internal/db/adapters"
	"github.com/moasq/go-b2b-starter/internal/db/postgres"
	adapterImpl "github.com/moasq/go-b2b-starter/internal/db/postgres/adapter_impl"
	sqlc "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
)

// Inject registers all database dependencies in the DI container
func Inject(container *dig.Container) error {
	// Register configuration
	if err := container.Provide(postgres.LoadConfig); err != nil {
		return fmt.Errorf("failed to provide database config: %w", err)
	}

	// Register connection pool
	if err := container.Provide(provideDBPool); err != nil {
		return fmt.Errorf("failed to provide database pool: %w", err)
	}

	// Register SQLC store
	if err := container.Provide(provideSQLCStore); err != nil {
		return fmt.Errorf("failed to provide SQLC store: %w", err)
	}

	// Register *sql.DB for modules that need standard database/sql interface
	if err := container.Provide(provideSQLDB); err != nil {
		return fmt.Errorf("failed to provide SQL DB: %w", err)
	}

	// Register domain stores
	if err := registerDomainStores(container); err != nil {
		return fmt.Errorf("failed to register domain stores: %w", err)
	}

	// Register database manager
	if err := container.Provide(provideDBManager); err != nil {
		return fmt.Errorf("failed to provide database manager: %w", err)
	}

	return nil
}

// provideDBPool creates the database connection pool
func provideDBPool(config postgres.Config) (*pgxpool.Pool, error) {
	return postgres.InitDB(config)
}

// provideSQLCStore creates the SQLC store
func provideSQLCStore(pool *pgxpool.Pool) sqlc.Store {
	return sqlc.NewStore(pool)
}

// provideSQLDB creates a *sql.DB from the pgxpool for compatibility
func provideSQLDB(pool *pgxpool.Pool) *sql.DB {
	// Use pgx stdlib to create a sql.DB from the pool connection string
	connConfig := pool.Config().ConnConfig
	return stdlib.OpenDB(*connConfig)
}

// provideDBManager creates the database manager for migrations and health checks
func provideDBManager(config postgres.Config, pool *pgxpool.Pool) *postgres.PostgresManager {
	return postgres.NewPostgresManager(config, pool)
}

// registerDomainStores registers all domain-specific repositories.
// These repositories implement domain ports using SQLC internally - no SQLC types leak out.
func registerDomainStores(container *dig.Container) error {
	providers := []any{
		func(store sqlc.Store) orgDomain.OrganizationRepository {
			return orgRepos.NewOrganizationRepository(store)
		},
		func(store sqlc.Store) orgDomain.AccountRepository { return orgRepos.NewAccountRepository(store) },
		func(store sqlc.Store) adapters.OrganizationStore { return adapterImpl.NewOrganizationStore(store) },
		func(store sqlc.Store) adapters.AccountStore { return adapterImpl.NewAccountStore(store) },
	}
	for _, provider := range providers {
		if err := container.Provide(provider); err != nil {
			return err
		}
	}
	return nil
}

// InjectWithOptions allows injecting with custom options
type InjectOptions struct {
	// SkipMigrations skips running database migrations
	SkipMigrations bool

	// SkipHealthCheck skips the initial health check
	SkipHealthCheck bool
}

// InjectWithOptions registers database dependencies with options
func InjectWithOptions(container *dig.Container, opts InjectOptions) error {
	if err := Inject(container); err != nil {
		return err
	}

	// Optionally run migrations and health checks
	if !opts.SkipMigrations || !opts.SkipHealthCheck {
		if err := container.Invoke(func(manager *postgres.PostgresManager) error {
			if !opts.SkipHealthCheck {
				if err := manager.CheckHealth(context.Background()); err != nil {
					return fmt.Errorf("database health check failed: %w", err)
				}
			}

			if !opts.SkipMigrations {
				if err := manager.RunMigrations(); err != nil {
					return fmt.Errorf("failed to run migrations: %w", err)
				}
			}

			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}
