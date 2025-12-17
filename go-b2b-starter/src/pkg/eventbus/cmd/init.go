package cmd

import "go.uber.org/dig"

func Init(container *dig.Container) error {
	if err := ProvideEventBus(container); err != nil {
		return err
	}
	
	return nil
}