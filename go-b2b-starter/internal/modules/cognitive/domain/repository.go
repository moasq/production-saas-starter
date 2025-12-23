package domain

import "context"

// EmbeddingRepository defines the interface for embedding data operations
type EmbeddingRepository interface {
	// Create creates a new document embedding
	Create(ctx context.Context, embedding *DocumentEmbedding) (*DocumentEmbedding, error)

	// GetByID retrieves an embedding by ID
	GetByID(ctx context.Context, orgID, embeddingID int32) (*DocumentEmbedding, error)

	// GetByDocumentID retrieves all embeddings for a document
	GetByDocumentID(ctx context.Context, orgID, documentID int32) ([]*DocumentEmbedding, error)

	// SearchSimilar finds similar documents using vector similarity
	SearchSimilar(ctx context.Context, orgID int32, embedding []float64, limit int32) ([]*SimilarDocument, error)

	// Delete removes embeddings for a document
	Delete(ctx context.Context, orgID, documentID int32) error

	// Count returns the total count of embeddings for an organization
	Count(ctx context.Context, orgID int32) (int64, error)
}

// ChatRepository defines the interface for chat session and message operations
type ChatRepository interface {
	// Sessions
	CreateSession(ctx context.Context, session *ChatSession) (*ChatSession, error)
	GetSessionByID(ctx context.Context, orgID, sessionID int32) (*ChatSession, error)
	ListSessionsByAccount(ctx context.Context, orgID, accountID int32, limit, offset int32) ([]*ChatSession, error)
	UpdateSessionTitle(ctx context.Context, orgID, sessionID int32, title string) (*ChatSession, error)
	DeleteSession(ctx context.Context, orgID, sessionID int32) error

	// Messages
	CreateMessage(ctx context.Context, message *ChatMessage) (*ChatMessage, error)
	GetMessagesBySession(ctx context.Context, sessionID int32) ([]*ChatMessage, error)
	GetRecentMessages(ctx context.Context, sessionID int32, limit int32) ([]*ChatMessage, error)
	CountMessagesBySession(ctx context.Context, sessionID int32) (int64, error)
	DeleteMessage(ctx context.Context, messageID int32) error
}
