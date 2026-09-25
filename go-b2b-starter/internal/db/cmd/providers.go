package cmd

import (
	"github.com/moasq/go-b2b-starter/internal/db"
	"go.uber.org/dig"
)

func ProvideDependencies(container *dig.Container) error { return db.Inject(container) }
