// Package cmd provides initialization for the users module.
package cmd

import (
	"fmt"

	"github.com/moasq/go-b2b-starter/internal/modules/users/domain"
	"github.com/moasq/go-b2b-starter/internal/modules/users/infra/repositories"
	"github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc"
	"go.uber.org/dig"
)

// Init initializes the users module.
//
// This sets up:
//   - domain.UserRepository
//   - domain.RefreshTokenRepository
//
// # Prerequisites
//
// The following must be available in the container:
//   - *sqlc.Queries (database queries)
func Init(container *dig.Container) error {
	// User Repository
	if err := container.Provide(func(queries *sqlc.Queries) domain.UserRepository {
		return repositories.NewUserRepository(queries)
	}); err != nil {
		return fmt.Errorf("failed to provide user repository: %w", err)
	}

	// Refresh Token Repository
	if err := container.Provide(func(queries *sqlc.Queries) domain.RefreshTokenRepository {
		return repositories.NewRefreshTokenRepository(queries)
	}); err != nil {
		return fmt.Errorf("failed to provide refresh token repository: %w", err)
	}

	return nil
}
