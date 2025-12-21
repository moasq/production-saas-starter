package postgres

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// InitDB initializes and returns a connection pool to the database
func InitDB(cfg Config) (*pgxpool.Pool, error) {

	// Create connection pool with retry logic
	var pool *pgxpool.Pool

	// Setup context with timeout for initial connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := RetryOperation(ctx, func(ctx context.Context) error {
		var connErr error
		pool, connErr = connPool(cfg)
		return connErr
	})

	if err != nil {
		log.Printf("Failed to initialize database connection after retries: %v", err)
		return nil, err
	}

	log.Println("Database connection successfully initialized")

	// Perform initial health check
	healthCtx, healthCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer healthCancel()

	manager := NewPostgresManager(cfg, pool)
	if err := manager.CheckHealth(healthCtx); err != nil {
		log.Printf("Initial database health check failed: %v", err)
		pool.Close()
		return nil, err
	}

	return pool, nil
}
