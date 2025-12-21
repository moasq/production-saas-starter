package postgres

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func connPool(cfg Config) (*pgxpool.Pool, error) {
	// Create a pool configuration
	poolConfig, err := pgxpool.ParseConfig(cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("unable to parse pool config: %w", err)
	}

	// Set pool configuration parameters
	poolConfig.MaxConns = int32(cfg.MaxConns)
	poolConfig.MinConns = int32(cfg.MinConns)
	poolConfig.MaxConnLifetime = cfg.ConnLifetime
	poolConfig.MaxConnIdleTime = cfg.ConnIdleTime
	poolConfig.HealthCheckPeriod = cfg.HealthCheckPeriod

	// Add connection lifecycle callbacks
	poolConfig.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
		// Optional validation before using a connection
		return true
	}

	poolConfig.AfterRelease = func(conn *pgx.Conn) bool {
		// Clean up after connection use if needed
		return true
	}

	// Create the connection pool with the configured settings
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Printf("Successfully connected to PostgreSQL database at %s:%s", cfg.Host, cfg.Port)

	return pool, nil
}
