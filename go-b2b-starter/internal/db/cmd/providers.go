package cmd

import (
	"github.com/moasq/go-b2b-starter/internal/db"
	"go.uber.org/dig"
)

// ProvideDependencies registers all database dependencies using the centralized inject
func ProvideDependencies(container *dig.Container) error {
	// Use the centralized inject function with default options
	return db.Inject(container)
}

// ProvideDependenciesWithOptions registers database dependencies with custom options
func ProvideDependenciesWithOptions(container *dig.Container, opts db.InjectOptions) error {
	return db.InjectWithOptions(container, opts)
}
