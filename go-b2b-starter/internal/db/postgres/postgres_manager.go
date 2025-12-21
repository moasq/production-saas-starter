package postgres

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"

	"context"

	"path/filepath"
	"strings"
	"time"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// PostgresManager implements DatabaseManager for PostgreSQL
type PostgresManager struct {
	config   Config
	connPool *pgxpool.Pool
}

func NewPostgresManager(config Config, connPool *pgxpool.Pool) *PostgresManager {
	return &PostgresManager{
		config:   config,
		connPool: connPool,
	}
}

// CheckHealth performs a health check on the database connection
func (pm *PostgresManager) CheckHealth(ctx context.Context) error {
	// Create a context with timeout for health check
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Try to ping the database
	err := pm.connPool.Ping(ctx)
	if err != nil {
		log.Printf("Database health check failed: %v", err)
		return fmt.Errorf("database health check failed: %w", err)
	}

	log.Println("Database health check successful")
	return nil
}

func (pm *PostgresManager) RunMigrations() error {
	ctx := context.Background()

	// Log the migration URL
	log.Printf("Migration URL: %s", pm.config.MigrationURL)

	// Check if the directory exists
	if _, err := os.Stat(pm.config.MigrationURL); os.IsNotExist(err) {
		log.Printf("Migration directory does not exist: %s", pm.config.MigrationURL)
		return fmt.Errorf("migration directory does not exist: %w", err)
	}

	files, err := os.ReadDir(pm.config.MigrationURL)
	if err != nil {
		log.Printf("Failed to read migrations directory: %v", err)
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	log.Printf("Total files in migration directory: %d", len(files))

	// Filter and sort migration files
	var migrationFiles []string
	for _, file := range files {
		log.Printf("Found file: %s", file.Name())
		if strings.HasSuffix(file.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}
	sort.Strings(migrationFiles)

	log.Printf("Migration files to execute: %v", migrationFiles)

	for _, fileName := range migrationFiles {
		fullPath := filepath.Join(pm.config.MigrationURL, fileName)
		log.Printf("Attempting to read file: %s", fullPath)

		content, err := os.ReadFile(fullPath)
		if err != nil {
			log.Printf("Failed to read migration file %s: %v", fileName, err)
			return fmt.Errorf("failed to read migration file %s: %w", fileName, err)
		}

		_, err = pm.connPool.Exec(ctx, string(content))
		if err != nil {
			log.Printf("Failed to execute migration file %s: %v", fileName, err)
			return fmt.Errorf("failed to execute migration file %s: %w", fileName, err)
		}

		log.Printf("Executed migration file: %s", fileName)
	}

	return nil
}

func (pm *PostgresManager) RunSeeds() error {
	ctx := context.Background()

	// Log the seed URL
	log.Printf("Seed URL: %s", pm.config.SeedURL)

	// Check if the directory exists
	if _, err := os.Stat(pm.config.SeedURL); os.IsNotExist(err) {
		log.Printf("Seed directory does not exist: %s", pm.config.SeedURL)
		return fmt.Errorf("seed directory does not exist: %w", err)
	}

	files, err := os.ReadDir(pm.config.SeedURL)
	if err != nil {
		log.Printf("Failed to read seeds directory: %v", err)
		return fmt.Errorf("failed to read seeds directory: %w", err)
	}

	log.Printf("Total files in seed directory: %d", len(files))

	for _, file := range files {
		log.Printf("Found file: %s", file.Name())
		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		fullPath := filepath.Join(pm.config.SeedURL, file.Name())
		log.Printf("Attempting to read file: %s", fullPath)

		content, err := os.ReadFile(fullPath)
		if err != nil {
			log.Printf("Failed to read seed file %s: %v", file.Name(), err)
			return fmt.Errorf("failed to read seed file %s: %w", file.Name(), err)
		}

		_, err = pm.connPool.Exec(ctx, string(content))
		if err != nil {
			log.Printf("Failed to execute seed file %s: %v", file.Name(), err)
			return fmt.Errorf("failed to execute seed file %s: %w", file.Name(), err)
		}

		log.Printf("Executed seed file: %s", file.Name())
	}

	return nil
}
