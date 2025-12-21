package cmd

import (
	"fmt"

	"github.com/moasq/go-b2b-starter/internal/polar"
	"go.uber.org/dig"
)

func Init(container *dig.Container) error {
	// Provide Polar configuration using viper
	if err := container.Provide(func() (*polar.Config, error) {
		config, err := polar.LoadConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load Polar configuration: %w", err)
		}

		return &config, nil
	}); err != nil {
		return fmt.Errorf("failed to provide Polar config: %w", err)
	}

	// Register Polar client
	if err := polar.Module(container); err != nil {
		return fmt.Errorf("failed to register Polar module: %w", err)
	}

	return nil
}
