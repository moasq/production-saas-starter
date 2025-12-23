// Package cmd provides initialization for the paywall module.
package cmd

import (
	"fmt"

	"github.com/moasq/go-b2b-starter/internal/modules/paywall"
	"go.uber.org/dig"
)

// InitMiddleware initializes the paywall middleware.
//
// This must be called after the billing module is initialized,
// as it depends on the SubscriptionStatusProvider from that module.
//
// # Prerequisites
//
// The following must be available in the container:
//   - paywall.SubscriptionStatusProvider (from app/billing module)
//
// # Usage
//
//	// After billing module init:
//	if err := paywallCmd.InitMiddleware(container); err != nil {
//	    panic(err)
//	}
func InitMiddleware(container *dig.Container) error {
	if err := paywall.SetupMiddleware(container); err != nil {
		return fmt.Errorf("failed to setup paywall middleware: %w", err)
	}
	return nil
}

// InitMiddlewareWithConfig initializes the paywall middleware with custom configuration.
//
// # Usage
//
//	config := &paywall.MiddlewareConfig{
//	    UpgradeURL: "/settings/billing",
//	}
//	if err := paywallCmd.InitMiddlewareWithConfig(container, config); err != nil {
//	    panic(err)
//	}
func InitMiddlewareWithConfig(container *dig.Container, config *paywall.MiddlewareConfig) error {
	if err := paywall.SetupMiddlewareWithConfig(container, config); err != nil {
		return fmt.Errorf("failed to setup paywall middleware: %w", err)
	}
	return nil
}

// SetupMiddleware is a direct alias to paywall.SetupMiddleware for convenience.
func SetupMiddleware(container *dig.Container) error {
	return paywall.SetupMiddleware(container)
}

// RegisterNamedMiddlewares is a direct alias to paywall.RegisterNamedMiddlewares for convenience.
func RegisterNamedMiddlewares(container *dig.Container) error {
	return paywall.RegisterNamedMiddlewares(container)
}
