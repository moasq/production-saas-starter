package polar

import (
	"fmt"

	"go.uber.org/dig"
)

// Module registers Polar package dependencies in the DI container
func Module(container *dig.Container) error {
	// Register Polar client
	if err := container.Provide(NewClient); err != nil {
		return fmt.Errorf("failed to provide Polar client: %w", err)
	}

	return nil
}
