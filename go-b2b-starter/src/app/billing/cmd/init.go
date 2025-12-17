package cmd

import (
	"go.uber.org/dig"
)

//
// The billing module handles subscription lifecycle management with Polar.sh:
//   - Webhook processing for subscription events
//   - Quota tracking and consumption
//   - Billing status queries
//
// Communication is event-driven:
//   - Polar sends webhook → billing processes event → updates local DB
//   - Paywall middleware reads from local DB (no external API calls)
func Init(container *dig.Container) error {
	// Register all dependencies
	if err := ProvideDependencies(container); err != nil {
		return err
	}

	return nil
}
