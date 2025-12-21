package services

import (
	"go.uber.org/dig"

	"github.com/moasq/go-b2b-starter/internal/billing/domain"
	"github.com/moasq/go-b2b-starter/internal/billing/infra/polar"
	"github.com/moasq/go-b2b-starter/internal/billing/infra/repositories"
	"github.com/moasq/go-b2b-starter/internal/db/adapters"
	logger "github.com/moasq/go-b2b-starter/internal/logger/domain"
	polarpkg "github.com/moasq/go-b2b-starter/internal/polar"
)

// Module handles dependency injection for billing services
type Module struct{}

func NewModule() *Module {
	return &Module{}
}

// Configure registers all services in the dependency container
func (m *Module) Configure(container *dig.Container) error {
	// Register SubscriptionRepository
	if err := container.Provide(func(store adapters.SubscriptionStore) domain.SubscriptionRepository {
		return repositories.NewSubscriptionRepository(store)
	}); err != nil {
		return err
	}

	// Register OrganizationAdapter
	if err := container.Provide(func(orgStore adapters.OrganizationStore) domain.OrganizationAdapter {
		return repositories.NewOrganizationAdapter(orgStore)
	}); err != nil {
		return err
	}

	// Register BillingProvider (Polar implementation)
	if err := container.Provide(func(client *polarpkg.Client) domain.BillingProvider {
		return polar.NewPolarAdapter(client)
	}); err != nil {
		return err
	}

	// Register BillingService
	if err := container.Provide(func(
		repo domain.SubscriptionRepository,
		orgAdapter domain.OrganizationAdapter,
		billingProvider domain.BillingProvider,
		logger logger.Logger,
	) BillingService {
		return NewBillingService(repo, orgAdapter, billingProvider, logger)
	}); err != nil {
		return err
	}

	return nil
}
