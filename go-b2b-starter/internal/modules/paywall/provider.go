package paywall

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/dig"
)

// ServerMiddlewareRegistrar is the interface for registering named middleware.
// This matches the server.Server interface's RegisterNamedMiddleware method.
type ServerMiddlewareRegistrar interface {
	RegisterNamedMiddleware(name string, middleware func() gin.HandlerFunc)
}

// SetupMiddleware wires the subscription middleware into the DI container.
//
// This must be called after the SubscriptionStatusProvider is available.
//
// # Prerequisites
//
// The following must be available in the container:
//   - subscription.SubscriptionStatusProvider
//
// # Usage
//
//	if err := subscription.SetupMiddleware(container); err != nil {
//	    return err
//	}
func SetupMiddleware(container *dig.Container) error {
	if err := container.Provide(func(
		provider SubscriptionStatusProvider,
	) *Middleware {
		return NewMiddleware(provider, nil)
	}); err != nil {
		return fmt.Errorf("failed to provide subscription middleware: %w", err)
	}

	return nil
}

// SetupMiddlewareWithConfig wires the subscription middleware with custom configuration.
//
// # Usage
//
//	config := &subscription.MiddlewareConfig{
//	    UpgradeURL: "/settings/billing",
//	}
//	if err := subscription.SetupMiddlewareWithConfig(container, config); err != nil {
//	    return err
//	}
func SetupMiddlewareWithConfig(container *dig.Container, config *MiddlewareConfig) error {
	if err := container.Provide(func(
		provider SubscriptionStatusProvider,
	) *Middleware {
		return NewMiddleware(provider, config)
	}); err != nil {
		return fmt.Errorf("failed to provide subscription middleware: %w", err)
	}

	return nil
}

// RegisterNamedMiddlewares registers the paywall middleware functions with the server.
//
// This should be called after SetupMiddleware and the server is available.
// It registers the following named middlewares:
//   - "paywall": RequireActiveSubscription middleware (blocks if inactive)
//   - "paywall_optional": OptionalSubscriptionStatus middleware (sets status, no blocking)
//
// For backward compatibility, these legacy names are also registered:
//   - "subscription" (deprecated, use "paywall")
//   - "subscription_optional" (deprecated, use "paywall_optional")
//
// # Usage
//
//	if err := paywall.RegisterNamedMiddlewares(container); err != nil {
//	    return err
//	}
func RegisterNamedMiddlewares(container *dig.Container) error {
	return container.Invoke(func(
		middleware *Middleware,
		server ServerMiddlewareRegistrar,
	) {
		// Register paywall middleware (requires active subscription)
		server.RegisterNamedMiddleware("paywall", func() gin.HandlerFunc {
			return middleware.RequireActiveSubscription()
		})

		// Register optional paywall middleware (sets status but doesn't block)
		server.RegisterNamedMiddleware("paywall_optional", func() gin.HandlerFunc {
			return middleware.OptionalSubscriptionStatus()
		})

		// Backward compatibility: legacy names (deprecated)
		server.RegisterNamedMiddleware("subscription", func() gin.HandlerFunc {
			return middleware.RequireActiveSubscription()
		})
		server.RegisterNamedMiddleware("subscription_optional", func() gin.HandlerFunc {
			return middleware.OptionalSubscriptionStatus()
		})
	})
}
