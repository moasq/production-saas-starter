package cmd

import (
	"github.com/moasq/go-b2b-starter/internal/modules/billing/app/services"
	"github.com/moasq/go-b2b-starter/internal/modules/billing/infra/polar"
	"github.com/moasq/go-b2b-starter/internal/modules/billing/infra/repositories"
	"go.uber.org/dig"
)

func Init(container *dig.Container) error {
	if err := container.Provide(repositories.NewOrganizationAdapter); err != nil {
		return err
	}
	if err := container.Provide(polar.NewPolarAdapter); err != nil {
		return err
	}
	return container.Provide(services.NewBillingService)
}
