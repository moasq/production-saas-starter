package cmd

import (
	"github.com/moasq/go-b2b-starter/internal/modules/billing/app/services"
	"go.uber.org/dig"
)

func ProvideDependencies(container *dig.Container) error {
	return services.NewModule().Configure(container)
}
