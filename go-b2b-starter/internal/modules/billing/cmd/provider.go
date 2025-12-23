package cmd

import (
	"fmt"

	"go.uber.org/dig"

	"github.com/moasq/go-b2b-starter/internal/modules/billing/app/services"
	"github.com/moasq/go-b2b-starter/internal/modules/billing/infra/adapters"
	"github.com/moasq/go-b2b-starter/internal/modules/paywall"
)

// ProvideDependencies registers all billing module dependencies
func ProvideDependencies(container *dig.Container) error {
	// Use the services module for dependency injection
	servicesModule := services.NewModule()
	if err := servicesModule.Configure(container); err != nil {
		return fmt.Errorf("failed to configure billing services: %w", err)
	}

	// Register SubscriptionStatusProvider for the paywall middleware
	// This adapter bridges the billing module to the pkg/paywall middleware
	// Communication is event-driven: webhooks → billing → DB → paywall reads
	if err := container.Provide(func(svc services.BillingService) paywall.SubscriptionStatusProvider {
		return adapters.NewStatusProviderAdapter(svc)
	}); err != nil {
		return fmt.Errorf("failed to provide subscription status provider: %w", err)
	}

	return nil
}
