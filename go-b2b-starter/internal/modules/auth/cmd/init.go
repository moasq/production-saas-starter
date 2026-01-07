// Package cmd provides initialization for the auth module.
package cmd

import (
	"fmt"

	"github.com/moasq/go-b2b-starter/internal/modules/auth"
	"github.com/moasq/go-b2b-starter/internal/modules/auth/adapters/local"
	"github.com/moasq/go-b2b-starter/internal/modules/users/domain"
	"github.com/moasq/go-b2b-starter/internal/platform/logger"
	"go.uber.org/dig"
)

// Init initializes the auth module with self-hosted JWT authentication.
//
// This sets up:
//   - local.Config (JWT and password configuration)
//   - auth.AuthProvider (local JWT adapter)
//   - local.Adapter (for login, register, etc.)
//
// # Prerequisites
//
// The following modules must be initialized first:
//   - users (for UserRepository and RefreshTokenRepository)
//   - logger
//
// # Usage
//
//	// In main/cmd/init_mods.go:
//	if err := authCmd.Init(container); err != nil {
//	    panic(err)
//	}
func Init(container *dig.Container) error {
	// Local auth configuration
	if err := container.Provide(func() (*local.Config, error) {
		return local.LoadConfig()
	}); err != nil {
		return fmt.Errorf("failed to provide local auth config: %w", err)
	}

	// Local Auth Adapter (implements auth.AuthProvider)
	if err := container.Provide(func(
		cfg *local.Config,
		userRepo domain.UserRepository,
		refreshRepo domain.RefreshTokenRepository,
		log logger.Logger,
	) (auth.AuthProvider, error) {
		adapter := local.NewAdapter(cfg, userRepo, refreshRepo, log)
		return adapter, nil
	}); err != nil {
		return fmt.Errorf("failed to provide auth provider: %w", err)
	}

	// Also provide the adapter directly for login/register operations
	if err := container.Provide(func(
		cfg *local.Config,
		userRepo domain.UserRepository,
		refreshRepo domain.RefreshTokenRepository,
		log logger.Logger,
	) *local.Adapter {
		return local.NewAdapter(cfg, userRepo, refreshRepo, log)
	}); err != nil {
		return fmt.Errorf("failed to provide local adapter: %w", err)
	}

	return nil
}

// InitMiddleware initializes the auth middleware.
//
// For B2C, this is simpler than B2B as we don't need organization resolvers.
//
// # Prerequisites
//
// The following must be available in the container:
//   - auth.AuthProvider (from Init)
//
// # Usage
//
//	if err := authCmd.InitMiddleware(container); err != nil {
//	    panic(err)
//	}
func InitMiddleware(container *dig.Container) error {
	if err := auth.SetupMiddleware(container); err != nil {
		return fmt.Errorf("failed to setup auth middleware: %w", err)
	}
	return nil
}

// RegisterNamedMiddlewares registers the auth middlewares with the server.
//
// This must be called after InitMiddleware and the server is available.
func RegisterNamedMiddlewares(container *dig.Container) error {
	return auth.RegisterNamedMiddlewares(container)
}
