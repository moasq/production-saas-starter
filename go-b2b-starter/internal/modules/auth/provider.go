package auth

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

// Provider handles dependency injection for the RBAC module
type Provider struct {
	container *dig.Container
}

func NewProvider(container *dig.Container) *Provider {
	return &Provider{
		container: container,
	}
}

// RegisterDependencies registers all RBAC dependencies in the container
func (p *Provider) RegisterDependencies() error {
	// Provide RBAC Service
	if err := p.container.Provide(func() RBACService {
		return NewRBACService()
	}); err != nil {
		return fmt.Errorf("failed to provide rbac service: %w", err)
	}

	// Provide RBAC Handler
	if err := p.container.Provide(func(service RBACService) *Handler {
		return NewHandler(service)
	}); err != nil {
		return fmt.Errorf("failed to provide rbac handler: %w", err)
	}

	// Provide RBAC Routes
	if err := p.container.Provide(func(handler *Handler) *Routes {
		return NewRoutes(handler)
	}); err != nil {
		return fmt.Errorf("failed to provide rbac routes: %w", err)
	}

	return nil
}

// SetupMiddleware wires the auth middleware into the DI container.
//
// This works for both B2B (with resolvers) and B2C (without resolvers) modes.
//
// # Prerequisites
//
// The following must be available in the container:
//   - auth.AuthProvider
//   - auth.OrganizationResolver (optional, for B2B)
//   - auth.AccountResolver (optional, for B2B)
//
// # Usage
//
//	if err := auth.SetupMiddleware(container); err != nil {
//	    return err
//	}
func SetupMiddleware(container *dig.Container) error {
	// For B2C mode, only require AuthProvider (no org/account resolvers)
	if err := container.Provide(func(
		provider AuthProvider,
	) *Middleware {
		return NewMiddleware(provider, nil, nil, nil)
	}); err != nil {
		return fmt.Errorf("failed to provide auth middleware: %w", err)
	}

	return nil
}

// RegisterNamedMiddlewares registers the auth middleware functions with the server.
//
// This should be called after SetupMiddleware and the server is available.
// It registers the following named middlewares:
//   - "auth": RequireAuth middleware (verifies JWT token)
//
// # Usage
//
//	if err := auth.RegisterNamedMiddlewares(container); err != nil {
//	    return err
//	}
func RegisterNamedMiddlewares(container *dig.Container) error {
	return container.Invoke(func(
		middleware *Middleware,
		server ServerMiddlewareRegistrar,
	) {
		// Register auth middleware (verifies JWT and sets Identity)
		server.RegisterNamedMiddleware("auth", func() gin.HandlerFunc {
			return middleware.RequireAuth()
		})
	})
}

// ProvideResolvers provides Organization and Account resolvers.
//
// This is a convenience function for when you have repositories that
// need to be adapted to the resolver interfaces.
//
// # Example
//
//	auth.ProvideResolvers(container, func(orgRepo domain.OrganizationRepository) auth.OrganizationResolver {
//	    return auth.NewOrganizationResolver(repo)
//	}, func(accRepo domain.AccountRepository) auth.AccountResolver {
//	    return auth.NewAccountResolver(repo)
//	})
func ProvideResolvers(
	container *dig.Container,
	orgResolverProvider any,
	accResolverProvider any,
) error {
	if err := container.Provide(orgResolverProvider); err != nil {
		return fmt.Errorf("failed to provide organization resolver: %w", err)
	}
	if err := container.Provide(accResolverProvider); err != nil {
		return fmt.Errorf("failed to provide account resolver: %w", err)
	}
	return nil
}
