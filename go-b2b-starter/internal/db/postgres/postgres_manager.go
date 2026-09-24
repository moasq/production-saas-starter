package postgres

import (
	"context"
	"embed"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"time"
)

//go:embed sqlc/migrations/*.sql archive/migrations/*.sql
var migrationFiles embed.FS

type PostgresManager struct {
	config   Config
	connPool *pgxpool.Pool
}

func NewPostgresManager(config Config, pool *pgxpool.Pool) *PostgresManager {
	return &PostgresManager{config, pool}
}
func (pm *PostgresManager) CheckHealth(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return pm.connPool.Ping(ctx)
}

// RunMigrations uses the migration library's database lock and dirty-version
// handling. SQL is embedded so the same binary deploys on every environment.
func (pm *PostgresManager) RunMigrations() error {
	active, err := iofs.New(migrationFiles, "sqlc/migrations")
	if err != nil {
		return err
	}
	legacy, err := iofs.New(migrationFiles, "archive/migrations")
	if err != nil {
		return err
	}
	src := &baselineSource{Driver: active, legacy: legacy}
	m, err := migrate.NewWithSourceInstance("iofs", src, pm.config.ConnectionString())
	if err != nil {
		return fmt.Errorf("open migrations: %w", err)
	}
	defer m.Close()
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate database: %w", err)
	}
	return nil
}

// baselineSource retains the immutable historical sources for migration ledger
// validation while applying only the new core baseline and subsequent migrations.
// It never runs old feature migrations on a fresh or partially upgraded database.
type baselineSource struct {
	source.Driver
	legacy source.Driver
}

func (s *baselineSource) Next(version uint) (uint, error) {
	if version < 10 {
		return s.Driver.First()
	}
	return s.Driver.Next(version)
}
func (s *baselineSource) ReadUp(version uint) (io.ReadCloser, string, error) {
	if version < 10 {
		return s.legacy.ReadUp(version)
	}
	return s.Driver.ReadUp(version)
}
func (s *baselineSource) ReadDown(version uint) (io.ReadCloser, string, error) {
	if version < 10 {
		return s.legacy.ReadDown(version)
	}
	return s.Driver.ReadDown(version)
}
func (s *baselineSource) Close() error {
	a := s.Driver.Close()
	b := s.legacy.Close()
	if a != nil {
		return a
	}
	return b
}
