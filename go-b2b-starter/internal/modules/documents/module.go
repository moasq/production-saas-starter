package documents

import (
	"go.uber.org/dig"

	"github.com/moasq/go-b2b-starter/internal/modules/documents/app/services"
	"github.com/moasq/go-b2b-starter/internal/modules/documents/domain"
	"github.com/moasq/go-b2b-starter/internal/platform/eventbus"
	filedomain "github.com/moasq/go-b2b-starter/internal/modules/files/domain"
	"github.com/moasq/go-b2b-starter/internal/platform/logger"
	ocrdomain "github.com/moasq/go-b2b-starter/internal/platform/ocr/domain"
)

// Module provides documents module dependencies
type Module struct {
	container *dig.Container
}

func NewModule(container *dig.Container) *Module {
	return &Module{
		container: container,
	}
}

// RegisterDependencies registers all documents module dependencies
// Note: Repository implementations are registered in internal/db/inject.go
func (m *Module) RegisterDependencies() error {
	// Register document service
	if err := m.container.Provide(func(
		docRepo domain.DocumentRepository,
		fileService filedomain.FileService,
		ocrService ocrdomain.OCRService,
		eventBus eventbus.EventBus,
		logger logger.Logger,
	) services.DocumentService {
		return services.NewDocumentService(docRepo, fileService, ocrService, eventBus, logger)
	}); err != nil {
		return err
	}

	return nil
}
