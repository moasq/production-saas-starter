package cmd

import (
	"log"

	"go.uber.org/dig"
	"github.com/moasq/go-b2b-starter/internal/modules/files/config"
)

func Init(container *dig.Container) {

	if err := container.Provide(config.LoadConfig); err != nil {
		log.Fatalf("Failed to provide file_manager config: %v", err)
	}

	SetupDependencies(container)
}
