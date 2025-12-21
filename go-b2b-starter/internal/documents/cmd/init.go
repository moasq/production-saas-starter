package cmd

import (
	"go.uber.org/dig"

	"github.com/moasq/go-b2b-starter/internal/documents"
)

func Init(container *dig.Container) error {
	module := documents.NewModule(container)
	return module.RegisterDependencies()
}
