package cmd

import (
	"fmt"
	"strings"

	"github.com/moasq/go-b2b-starter/internal/platform/logger"
	"github.com/moasq/go-b2b-starter/internal/platform/redis"
	"github.com/moasq/go-b2b-starter/internal/platform/stytch"
	"go.uber.org/dig"
)

// ProvideStytchDependencies wires the Stytch configuration and client into the DI container.
func ProvideStytchDependencies(container *dig.Container) error {
	providers := []any{
		stytch.LoadConfig,
		provideStytchClient,
		provideRBACPolicyService,
	}

	for _, provider := range providers {
		if err := container.Provide(provider); err != nil {
			return fmt.Errorf("failed to provide Stytch dependency: %w", err)
		}
	}

	return nil
}

func provideStytchClient(cfg *stytch.Config, log logger.Logger) (*stytch.Client, error) {
	// Check for placeholder credentials (development mode)
	if isPlaceholderCredentials(cfg) {
		log.Warn("Stytch credentials are placeholders - Stytch client will be nil (development mode)", map[string]any{
			"project_id": cfg.ProjectID,
			"message":    "Organization/member management features will not work. Update STYTCH_PROJECT_ID and STYTCH_SECRET in app.env",
		})
		// Return nil client for development mode - app/auth repositories should handle nil gracefully
		return nil, nil
	}

	return stytch.NewClient(*cfg)
}

// isPlaceholderCredentials checks if the Stytch credentials are placeholder values.
func isPlaceholderCredentials(cfg *stytch.Config) bool {
	return strings.Contains(cfg.ProjectID, "REPLACE") ||
		strings.Contains(cfg.Secret, "REPLACE") ||
		cfg.ProjectID == "" ||
		cfg.Secret == ""
}

func provideRBACPolicyService(
	client *stytch.Client,
	redisClient redis.Client,
	log logger.Logger,
) *stytch.RBACPolicyService {
	// If client is nil (development mode), return nil for RBAC service too
	if client == nil {
		return nil
	}
	return stytch.NewRBACPolicyService(client, redisClient, log)
}
