package cmd

import (
	"log"

	"go.uber.org/dig"
)

func Init(dig *dig.Container) error {
	if err := provideRedisDependencies(dig); err != nil {
		log.Fatalf("Failed to provide Redis dependencies: %v", err)
		return err
	}
	return nil
}
