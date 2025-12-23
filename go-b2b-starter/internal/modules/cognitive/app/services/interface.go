package services

import (
	"context"

	"github.com/moasq/go-b2b-starter/internal/modules/cognitive/domain"
)

// EmbeddingService defines the interface for embedding operations
type EmbeddingService interface {
	// EmbedDocument generates and stores embeddings for a document
	EmbedDocument(ctx context.Context, orgID, documentID int32, text string) (*domain.DocumentEmbedding, error)

	// GetDocumentEmbeddings retrieves embeddings for a document
	GetDocumentEmbeddings(ctx context.Context, orgID, documentID int32) ([]*domain.DocumentEmbedding, error)

	// SearchSimilarDocuments finds documents similar to the given text
	SearchSimilarDocuments(ctx context.Context, orgID int32, text string, limit int32) ([]*domain.SimilarDocument, error)

	// DeleteDocumentEmbeddings removes embeddings for a document
	DeleteDocumentEmbeddings(ctx context.Context, orgID, documentID int32) error

	// GetStats retrieves embedding statistics
	GetStats(ctx context.Context, orgID int32) (*domain.EmbeddingStats, error)
}

// RAGService defines the interface for RAG (Retrieval-Augmented Generation) operations
type RAGService interface {
	// Chat sends a message and gets a response, optionally using RAG
	Chat(ctx context.Context, orgID, accountID int32, req *domain.ChatRequest) (*domain.ChatResponse, error)

	// GetSession retrieves a chat session
	GetSession(ctx context.Context, orgID, sessionID int32) (*domain.ChatSession, error)

	// ListSessions lists chat sessions for an account
	ListSessions(ctx context.Context, orgID, accountID int32, limit, offset int32) ([]*domain.ChatSession, error)

	// DeleteSession deletes a chat session
	DeleteSession(ctx context.Context, orgID, sessionID int32) error

	// GetSessionHistory retrieves messages for a session
	GetSessionHistory(ctx context.Context, orgID, sessionID int32) ([]*domain.ChatMessage, error)

	// UpdateSessionTitle updates the title of a chat session
	UpdateSessionTitle(ctx context.Context, orgID, sessionID int32, title string) (*domain.ChatSession, error)
}

// DocumentListener handles document events from the documents module
type DocumentListener interface {
	// HandleDocumentUploaded processes the DocumentUploaded event
	HandleDocumentUploaded(ctx context.Context, documentID, orgID int32, text string) error
}
