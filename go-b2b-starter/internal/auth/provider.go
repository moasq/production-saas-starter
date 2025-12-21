package rbac

import (
	"fmt"

	"github.com/moasq/go-b2b-starter/internal/auth"
	"go.uber.org/dig"
)

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
	if err := p.container.Provide(func() auth.RBACService {
		return auth.NewRBACService()
	}); err != nil {
		return fmt.Errorf("failed to provide rbac service: %w", err)
	}

	// Provide RBAC Handler
	if err := p.container.Provide(func(service auth.RBACService) *Handler {
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
