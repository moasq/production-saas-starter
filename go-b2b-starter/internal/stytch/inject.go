package stytch

import (
	"fmt"

	"github.com/moasq/go-b2b-starter/internal/logger"
	"github.com/moasq/go-b2b-starter/internal/redis"
	"go.uber.org/dig"
)

// ProvideDependencies registers Stytch package dependencies in the DI container
func ProvideDependencies(container *dig.Container) error {
	// Provide Stytch client
	if err := container.Provide(func(cfg Config) (*Client, error) {
		return NewClient(cfg)
	}); err != nil {
		return fmt.Errorf("failed to provide stytch client: %w", err)
	}

	// Provide RBAC policy service
	if err := container.Provide(func(
		client *Client,
		redisClient redis.Client,
		logger logger.Logger,
	) *RBACPolicyService {
		return NewRBACPolicyService(client, redisClient, logger)
	}); err != nil {
		return fmt.Errorf("failed to provide RBAC policy service: %w", err)
	}

	return nil
}
