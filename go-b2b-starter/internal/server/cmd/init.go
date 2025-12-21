package cmd

import "go.uber.org/dig"

func Init(container *dig.Container) {

	SetupDependencies(container)

}
