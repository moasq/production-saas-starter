package cmd

import (
	"fmt"

	"github.com/moasq/go-b2b-starter/internal/platform/redis"
	"go.uber.org/dig"
)

func provideRedisDependencies(container *dig.Container) error {
	providers := []any{
		redis.LoadConfig,
		provideRedisStore,
	}

	for _, provider := range providers {
		if err := container.Provide(provider); err != nil {
			return fmt.Errorf("failed to provide Redis dependency: %w", err)
		}
	}

	return nil
}

func provideRedisStore() (redis.Client, error) {
	return redis.InitRedis()
}
