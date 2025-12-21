package cmd

import (
	"go.uber.org/dig"

	"github.com/moasq/go-b2b-starter/internal/ocr/domain"
	"github.com/moasq/go-b2b-starter/internal/ocr/infra"
	loggerDomain "github.com/moasq/go-b2b-starter/internal/logger/domain"
)

func Init(container *dig.Container) error {
	return container.Provide(func(logger loggerDomain.Logger) (domain.OCRService, error) {
		config := infra.NewOCRConfig()
		return infra.NewMistralOCRClient(config, logger)
	})
}