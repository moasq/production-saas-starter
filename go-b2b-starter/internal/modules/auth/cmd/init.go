// Package cmd provides initialization for the auth module.
package cmd

import (
	"fmt"
	"strings"

	"github.com/moasq/go-b2b-starter/internal/modules/auth"
	"github.com/moasq/go-b2b-starter/internal/modules/auth/adapters/stytch"
	"github.com/moasq/go-b2b-starter/internal/platform/logger"
	"github.com/moasq/go-b2b-starter/internal/platform/redis"
	"go.uber.org/dig"
)

//
// This sets up:
//   - stytch.Config
//   - auth.AuthProvider (Stytch adapter)
//
// Note: The auth middleware is NOT initialized here because it requires
// organization/account resolvers from the organizations module.
// Use InitMiddleware after the organizations module is initialized.
//
// # Prerequisites
//
// The following modules must be initialized first:
//   - redis (for caching)
//   - logger
//
// # Usage
//
//	// In main/cmd/init_mods.go:
//	if err := authCmd.Init(container); err != nil {
//	    panic(err)
//	}
func Init(container *dig.Container) error {
	// Stytch configuration
	if err := container.Provide(func() (*stytch.Config, error) {
		return stytch.LoadConfig()
	}); err != nil {
		return fmt.Errorf("failed to provide stytch config: %w", err)
	}

	// Stytch Auth Adapter (implements auth.AuthProvider)
	if err := container.Provide(func(
		cfg *stytch.Config,
		redisClient redis.Client,
		log logger.Logger,
	) (auth.AuthProvider, error) {
		// Check for placeholder credentials
		if isPlaceholderCredentials(cfg) {
			log.Warn("Stytch credentials are placeholders - using development mode", map[string]any{
				"project_id": cfg.ProjectID,
				"message":    "Update STYTCH_PROJECT_ID and STYTCH_SECRET in app.env with real credentials",
			})
			return stytch.NewMockAuthAdapter(log), nil
		}

		adapter, err := stytch.NewStytchAuthAdapter(cfg, redisClient, log)
		if err != nil {
			return nil, fmt.Errorf("failed to create stytch adapter: %w", err)
		}
		return adapter, nil
	}); err != nil {
		return fmt.Errorf("failed to provide auth provider: %w", err)
	}

	return nil
}

// InitMiddleware initializes the auth middleware with resolvers.
//
// This must be called after the organizations module is initialized,
// as it depends on organization and account repositories.
//
// # Prerequisites
//
// The following must be available in the container:
//   - auth.AuthProvider (from Init)
//   - auth.OrganizationResolver
//   - auth.AccountResolver
//   - serverDomain.Server (for registering named middlewares)
//
// # Usage
//
//	// After organizations module init:
//	if err := authCmd.InitMiddleware(container); err != nil {
//	    panic(err)
//	}
func InitMiddleware(container *dig.Container) error {
	if err := auth.SetupMiddleware(container); err != nil {
		return fmt.Errorf("failed to setup auth middleware: %w", err)
	}
	return nil
}

// isPlaceholderCredentials checks if the Stytch credentials are placeholder values.
func isPlaceholderCredentials(cfg *stytch.Config) bool {
	return strings.Contains(cfg.ProjectID, "REPLACE") ||
		strings.Contains(cfg.Secret, "REPLACE") ||
		cfg.ProjectID == "" ||
		cfg.Secret == ""
}
