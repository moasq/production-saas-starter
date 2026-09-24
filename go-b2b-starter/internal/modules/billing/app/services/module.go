package services

import (
	"github.com/moasq/go-b2b-starter/internal/db/adapters"
	"github.com/moasq/go-b2b-starter/internal/modules/billing/domain"
	"github.com/moasq/go-b2b-starter/internal/modules/billing/infra/polar"
	"github.com/moasq/go-b2b-starter/internal/modules/billing/infra/repositories"
	"go.uber.org/dig"
)

type Module struct{}

func NewModule() *Module { return &Module{} }
func (m *Module) Configure(container *dig.Container) error {
	if err := container.Provide(func(store adapters.OrganizationStore) domain.OrganizationAdapter {
		return repositories.NewOrganizationAdapter(store)
	}); err != nil {
		return err
	}
	if err := container.Provide(polar.NewPolarAdapter); err != nil {
		return err
	}
	return container.Provide(NewBillingService)
}
