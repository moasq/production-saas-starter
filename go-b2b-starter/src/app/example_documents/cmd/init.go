package cmd

import (
	"go.uber.org/dig"

	"github.com/moasq/go-b2b-starter/app/example_documents"
)

func Init(container *dig.Container) error {
	module := documents.NewModule(container)
	return module.RegisterDependencies()
}
