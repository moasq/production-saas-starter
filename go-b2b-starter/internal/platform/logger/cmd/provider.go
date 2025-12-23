package cmd

import (
	"github.com/moasq/go-b2b-starter/internal/platform/logger"
	"go.uber.org/dig"
)

func ProvideDependencies(container *dig.Container) {
	container.Provide(logger.New)
}
