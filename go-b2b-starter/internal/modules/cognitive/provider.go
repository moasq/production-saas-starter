package cognitive

import (
	"go.uber.org/dig"
)

type Provider struct {
	container *dig.Container
}

func NewProvider(container *dig.Container) *Provider {
	return &Provider{container: container}
}

func (p *Provider) RegisterDependencies() error {
	// Register handler
	if err := p.container.Provide(NewHandler); err != nil {
		return err
	}

	// Register routes
	if err := p.container.Provide(NewRoutes); err != nil {
		return err
	}

	return nil
}
