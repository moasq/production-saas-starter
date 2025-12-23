package cmd

import (
	"go.uber.org/dig"

	"github.com/moasq/go-b2b-starter/internal/platform/llm/domain"
	"github.com/moasq/go-b2b-starter/internal/platform/llm/infra"
	loggerDomain "github.com/moasq/go-b2b-starter/internal/platform/logger/domain"
)

func Init(container *dig.Container) error {
	// Register LLMClient (which includes LLMService)
	if err := container.Provide(func(logger loggerDomain.Logger) (domain.LLMClient, error) {
		config := infra.NewLLMConfig()
		return infra.NewOpenAIClient(config, logger)
	}); err != nil {
		return err
	}

	// Also register LLMService for backward compatibility
	return container.Provide(func(client domain.LLMClient) domain.LLMService {
		return client
	})
}