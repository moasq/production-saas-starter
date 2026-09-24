package billing

import (
	"go.uber.org/dig"
)

// RegisterHandlers registers subscription API handlers in the DI container
func RegisterHandlers(container *dig.Container) error {
	return container.Provide(NewHandler)
}
