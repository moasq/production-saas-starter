package cognitive

import (
	"go.uber.org/dig"

	"github.com/moasq/go-b2b-starter/internal/modules/cognitive/app/services"
	"github.com/moasq/go-b2b-starter/internal/modules/cognitive/domain"
	"github.com/moasq/go-b2b-starter/internal/modules/cognitive/infra/ai"
	llmdomain "github.com/moasq/go-b2b-starter/internal/platform/llm/domain"
)

// Module provides cognitive module dependencies
type Module struct {
	container *dig.Container
}

func NewModule(container *dig.Container) *Module {
	return &Module{
		container: container,
	}
}

// RegisterDependencies registers all cognitive module dependencies
// Note: Repository implementations are registered in internal/db/inject.go
func (m *Module) RegisterDependencies() error {
	// Register AI adapters (infra layer)
	if err := m.container.Provide(func(
		llmClient llmdomain.LLMClient,
	) domain.TextVectorizer {
		return ai.NewTextVectorizer(llmClient)
	}); err != nil {
		return err
	}

	if err := m.container.Provide(func(
		llmClient llmdomain.LLMClient,
	) domain.AssistantProvider {
		return ai.NewAssistantProvider(llmClient)
	}); err != nil {
		return err
	}

	// Register embedding service
	if err := m.container.Provide(func(
		embeddingRepo domain.EmbeddingRepository,
		textVectorizer domain.TextVectorizer,
	) services.EmbeddingService {
		return services.NewEmbeddingService(embeddingRepo, textVectorizer)
	}); err != nil {
		return err
	}

	// Register RAG service
	if err := m.container.Provide(func(
		chatRepo domain.ChatRepository,
		embeddingRepo domain.EmbeddingRepository,
		textVectorizer domain.TextVectorizer,
		assistantProvider domain.AssistantProvider,
	) services.RAGService {
		return services.NewRAGService(chatRepo, embeddingRepo, textVectorizer, assistantProvider)
	}); err != nil {
		return err
	}

	// Register document listener
	if err := m.container.Provide(func(
		embeddingService services.EmbeddingService,
	) services.DocumentListener {
		return services.NewDocumentListener(embeddingService)
	}); err != nil {
		return err
	}

	return nil
}
