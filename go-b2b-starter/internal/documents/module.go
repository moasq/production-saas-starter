package documents

import (
	"go.uber.org/dig"

	"github.com/moasq/go-b2b-starter/internal/documents/app/services"
	"github.com/moasq/go-b2b-starter/internal/documents/domain"
	"github.com/moasq/go-b2b-starter/internal/documents/infra/repositories"
	"github.com/moasq/go-b2b-starter/internal/db/adapters"
	"github.com/moasq/go-b2b-starter/internal/eventbus"
	filedomain "github.com/moasq/go-b2b-starter/internal/files/domain"
	"github.com/moasq/go-b2b-starter/internal/logger"
	ocrdomain "github.com/moasq/go-b2b-starter/internal/ocr/domain"
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
func (m *Module) RegisterDependencies() error {
	// Register document repository
	if err := m.container.Provide(func(
		docStore adapters.DocumentStore,
	) domain.DocumentRepository {
		return repositories.NewDocumentRepository(docStore)
	}); err != nil {
		return err
	}

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
