package ai

import (
	"context"

	"github.com/moasq/go-b2b-starter/app/example_cognitive/domain"
	llmdomain "github.com/moasq/go-b2b-starter/pkg/llm/domain"
)

const embeddingModel = "text-embedding-3-small"

type openAITextVectorizer struct {
	llmClient llmdomain.LLMClient
}

func NewTextVectorizer(llmClient llmdomain.LLMClient) domain.TextVectorizer {
	return &openAITextVectorizer{llmClient: llmClient}
}

func (v *openAITextVectorizer) Vectorize(ctx context.Context, text string) ([]float64, error) {
	return v.llmClient.GenerateEmbedding(ctx, text, embeddingModel)
}
