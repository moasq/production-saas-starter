package domain

import "errors"

// Domain errors for cognitive module
var (
	// Embedding errors
	ErrEmbeddingNotFound         = errors.New("embedding not found")
	ErrEmbeddingGenerationFailed = errors.New("failed to generate embedding")
	ErrEmbeddingAlreadyExists    = errors.New("embedding already exists for this document")

	// Session errors
	ErrSessionNotFound             = errors.New("chat session not found")
	ErrSessionOrganizationRequired = errors.New("session organization ID is required")
	ErrSessionAccountRequired      = errors.New("session account ID is required")

	// Message errors
	ErrMessageNotFound        = errors.New("chat message not found")
	ErrMessageSessionRequired = errors.New("message session ID is required")
	ErrMessageContentRequired = errors.New("message content is required")
	ErrMessageRoleRequired    = errors.New("message role is required")

	// RAG errors
	ErrRAGContextEmpty      = errors.New("no relevant documents found for RAG context")
	ErrRAGSearchFailed      = errors.New("RAG similarity search failed")
	ErrRAGCompletionFailed  = errors.New("RAG completion generation failed")

	// LLM errors
	ErrLLMUnavailable      = errors.New("LLM service is unavailable")
	ErrLLMRequestFailed    = errors.New("LLM request failed")
	ErrLLMResponseInvalid  = errors.New("LLM response is invalid")
)
