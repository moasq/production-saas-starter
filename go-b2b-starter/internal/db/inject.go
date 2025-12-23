package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/dig"

	// Domain interfaces - these are the interfaces we provide
	billingDomain "github.com/moasq/go-b2b-starter/internal/modules/billing/domain"
	cognitiveDomain "github.com/moasq/go-b2b-starter/internal/modules/cognitive/domain"
	documentDomain "github.com/moasq/go-b2b-starter/internal/modules/documents/domain"
	fileDomain "github.com/moasq/go-b2b-starter/internal/modules/files/domain"
	orgDomain "github.com/moasq/go-b2b-starter/internal/modules/organizations/domain"

	// Repository implementations from module infra layers
	billingRepos "github.com/moasq/go-b2b-starter/internal/modules/billing/infra/repositories"
	cognitiveRepos "github.com/moasq/go-b2b-starter/internal/modules/cognitive/infra/repositories"
	documentRepos "github.com/moasq/go-b2b-starter/internal/modules/documents/infra/repositories"
	fileInfra "github.com/moasq/go-b2b-starter/internal/modules/files/infra"
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
	// ============================================
	// NEW: Sealed repository implementations
	// These use domain interfaces and hide SQLC internals
	// ============================================

	// Register DocumentRepository - implements documents/domain.DocumentRepository
	if err := container.Provide(func(sqlcStore sqlc.Store) documentDomain.DocumentRepository {
		return documentRepos.NewDocumentRepository(sqlcStore)
	}); err != nil {
		return fmt.Errorf("failed to provide document repository: %w", err)
	}

	// Register OrganizationRepository - implements organizations/domain.OrganizationRepository
	if err := container.Provide(func(sqlcStore sqlc.Store) orgDomain.OrganizationRepository {
		return orgRepos.NewOrganizationRepository(sqlcStore)
	}); err != nil {
		return fmt.Errorf("failed to provide organization repository: %w", err)
	}

	// Register AccountRepository - implements organizations/domain.AccountRepository
	if err := container.Provide(func(sqlcStore sqlc.Store) orgDomain.AccountRepository {
		return orgRepos.NewAccountRepository(sqlcStore)
	}); err != nil {
		return fmt.Errorf("failed to provide account repository: %w", err)
	}

	// Register SubscriptionRepository - implements billing/domain.SubscriptionRepository
	if err := container.Provide(func(sqlcStore sqlc.Store) billingDomain.SubscriptionRepository {
		return billingRepos.NewSubscriptionRepository(sqlcStore)
	}); err != nil {
		return fmt.Errorf("failed to provide subscription repository: %w", err)
	}

	// Register EmbeddingRepository - implements cognitive/domain.EmbeddingRepository
	if err := container.Provide(func(sqlcStore sqlc.Store) cognitiveDomain.EmbeddingRepository {
		return cognitiveRepos.NewEmbeddingRepository(sqlcStore)
	}); err != nil {
		return fmt.Errorf("failed to provide embedding repository: %w", err)
	}

	// Register ChatRepository - implements cognitive/domain.ChatRepository
	if err := container.Provide(func(sqlcStore sqlc.Store) cognitiveDomain.ChatRepository {
		return cognitiveRepos.NewChatRepository(sqlcStore)
	}); err != nil {
		return fmt.Errorf("failed to provide chat repository: %w", err)
	}

	// Register FileMetadataRepository - implements files/domain.FileMetadataRepository
	if err := container.Provide(func(sqlcStore sqlc.Store) fileDomain.FileMetadataRepository {
		return fileInfra.NewFileMetadataRepository(sqlcStore)
	}); err != nil {
		return fmt.Errorf("failed to provide file metadata repository: %w", err)
	}

	// ============================================
	// LEGACY: Adapter stores (kept for backward compatibility)
	// TODO: Migrate callers to use domain interfaces, then remove these
	// ============================================

	// Register FileAssetStore - thin wrapper for file management operations
	if err := container.Provide(func(sqlcStore sqlc.Store) adapters.FileAssetStore {
		return adapterImpl.NewFileAssetStore(sqlcStore)
	}); err != nil {
		return fmt.Errorf("failed to provide file asset store: %w", err)
	}

	// Register OrganizationStore - thin wrapper for organization operations
	if err := container.Provide(func(sqlcStore sqlc.Store) adapters.OrganizationStore {
		return adapterImpl.NewOrganizationStore(sqlcStore)
	}); err != nil {
		return fmt.Errorf("failed to provide organization store: %w", err)
	}

	// Register AccountStore - thin wrapper for account operations
	if err := container.Provide(func(sqlcStore sqlc.Store) adapters.AccountStore {
		return adapterImpl.NewAccountStore(sqlcStore)
	}); err != nil {
		return fmt.Errorf("failed to provide account store: %w", err)
	}

	// Register SubscriptionStore - thin wrapper for subscription billing operations
	if err := container.Provide(func(sqlcStore sqlc.Store) adapters.SubscriptionStore {
		return adapterImpl.NewSubscriptionStore(sqlcStore)
	}); err != nil {
		return fmt.Errorf("failed to provide subscription store: %w", err)
	}

	// Register DocumentStore - thin wrapper for document operations
	if err := container.Provide(func(sqlcStore sqlc.Store) adapters.DocumentStore {
		return adapterImpl.NewDocumentStore(sqlcStore)
	}); err != nil {
		return fmt.Errorf("failed to provide document store: %w", err)
	}

	// Register EmbeddingStore - thin wrapper for cognitive embedding operations
	if err := container.Provide(func(sqlcStore sqlc.Store) adapters.EmbeddingStore {
		return adapterImpl.NewEmbeddingStore(sqlcStore)
	}); err != nil {
		return fmt.Errorf("failed to provide embedding store: %w", err)
	}

	// Register ChatStore - thin wrapper for cognitive chat operations
	if err := container.Provide(func(sqlcStore sqlc.Store) adapters.ChatStore {
		return adapterImpl.NewChatStore(sqlcStore)
	}); err != nil {
		return fmt.Errorf("failed to provide chat store: %w", err)
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
