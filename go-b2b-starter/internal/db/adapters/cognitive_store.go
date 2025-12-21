package adapters

import (
	"context"

	db "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
	"github.com/pgvector/pgvector-go"
)

// EmbeddingStore provides database operations for document embeddings
type EmbeddingStore interface {
	CreateDocumentEmbedding(ctx context.Context, arg db.CreateDocumentEmbeddingParams) (db.CognitiveDocumentEmbedding, error)
	GetDocumentEmbeddingByID(ctx context.Context, arg db.GetDocumentEmbeddingByIDParams) (db.CognitiveDocumentEmbedding, error)
	GetDocumentEmbeddingsByDocumentID(ctx context.Context, arg db.GetDocumentEmbeddingsByDocumentIDParams) ([]db.CognitiveDocumentEmbedding, error)
	SearchSimilarDocuments(ctx context.Context, arg db.SearchSimilarDocumentsParams) ([]db.SearchSimilarDocumentsRow, error)
	DeleteDocumentEmbeddings(ctx context.Context, arg db.DeleteDocumentEmbeddingsParams) error
	CountDocumentEmbeddingsByOrganization(ctx context.Context, organizationID int32) (int64, error)
}

// ChatStore provides database operations for chat sessions and messages
type ChatStore interface {
	// Sessions
	CreateChatSession(ctx context.Context, arg db.CreateChatSessionParams) (db.CognitiveChatSession, error)
	GetChatSessionByID(ctx context.Context, arg db.GetChatSessionByIDParams) (db.CognitiveChatSession, error)
	ListChatSessionsByAccount(ctx context.Context, arg db.ListChatSessionsByAccountParams) ([]db.CognitiveChatSession, error)
	UpdateChatSessionTitle(ctx context.Context, arg db.UpdateChatSessionTitleParams) (db.CognitiveChatSession, error)
	DeleteChatSession(ctx context.Context, arg db.DeleteChatSessionParams) error

	// Messages
	CreateChatMessage(ctx context.Context, arg db.CreateChatMessageParams) (db.CognitiveChatMessage, error)
	GetChatMessagesBySession(ctx context.Context, sessionID int32) ([]db.CognitiveChatMessage, error)
	GetRecentChatMessages(ctx context.Context, arg db.GetRecentChatMessagesParams) ([]db.CognitiveChatMessage, error)
	CountChatMessagesBySession(ctx context.Context, sessionID int32) (int64, error)
	DeleteChatMessage(ctx context.Context, id int32) error
}

// VectorHelper provides utilities for working with pgvector
type VectorHelper interface {
	ToVector(embedding []float64) pgvector.Vector
	FromVector(v pgvector.Vector) []float64
}
