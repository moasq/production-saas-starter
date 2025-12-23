package domain

import "context"

// TextVectorizer creates searchable vector representations of text content.
// This enables semantic document search and similarity matching.
// Implementation details (embedding models, providers) are in the infra layer.
type TextVectorizer interface {
	// Vectorize converts text content into a searchable vector representation
	Vectorize(ctx context.Context, text string) ([]float64, error)
}

// AssistantProvider provides AI-powered conversational assistance.
// This enables intelligent responses based on context and user queries.
// Implementation details (LLM providers, models) are in the infra layer.
type AssistantProvider interface {
	// GenerateResponse creates an AI response for the given prompt with context
	GenerateResponse(ctx context.Context, prompt string) (*AssistantResponse, error)
}

// AssistantResponse contains the result of an AI assistance request
type AssistantResponse struct {
	Content    string // The generated response text
	TokensUsed int    // Tokens consumed (for usage tracking)
}
