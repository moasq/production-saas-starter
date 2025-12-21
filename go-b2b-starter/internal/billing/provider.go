package subscriptions

import (
	"go.uber.org/dig"
)

// RegisterHandlers registers subscription API handlers in the DI container
func RegisterHandlers(container *dig.Container) error {
	if err := container.Provide(NewHandler); err != nil {
		return err
	}
	return nil
}

// ProvideHandler is an alias for RegisterHandlers for consistency
func ProvideHandler(container *dig.Container) error {
	return RegisterHandlers(container)
}
