package domain

import "context"

type CompletionRequest struct {
	Prompt      string
	MaxTokens   *int
	Temperature *float32
}

type CompletionResponse struct {
	Text       string
	TokensUsed int
	Model      string
}

type EmbeddingRequest struct {
	Text  string
	Model string
}

type EmbeddingResponse struct {
	Embedding  []float64
	TokensUsed int
	Model      string
}

type StreamChunk struct {
	Content string
	Done    bool
}

type LLMService interface {
	Complete(ctx context.Context, request CompletionRequest) (*CompletionResponse, error)
	CompleteStream(ctx context.Context, request CompletionRequest, callback func(StreamChunk) error) (*CompletionResponse, error)
}

type LLMClient interface {
	LLMService
	GenerateEmbedding(ctx context.Context, text string, model string) ([]float64, error)
}