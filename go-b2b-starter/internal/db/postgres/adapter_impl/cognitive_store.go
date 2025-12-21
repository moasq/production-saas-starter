package adapterimpl

import (
	"context"

	"github.com/moasq/go-b2b-starter/internal/db/adapters"
	sqlc "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
)

// embeddingStore implements adapters.EmbeddingStore
type embeddingStore struct {
	store sqlc.Store
}

func NewEmbeddingStore(store sqlc.Store) adapters.EmbeddingStore {
	return &embeddingStore{store: store}
}

func (s *embeddingStore) CreateDocumentEmbedding(ctx context.Context, arg sqlc.CreateDocumentEmbeddingParams) (sqlc.CognitiveDocumentEmbedding, error) {
	return s.store.CreateDocumentEmbedding(ctx, arg)
}

func (s *embeddingStore) GetDocumentEmbeddingByID(ctx context.Context, arg sqlc.GetDocumentEmbeddingByIDParams) (sqlc.CognitiveDocumentEmbedding, error) {
	return s.store.GetDocumentEmbeddingByID(ctx, arg)
}

func (s *embeddingStore) GetDocumentEmbeddingsByDocumentID(ctx context.Context, arg sqlc.GetDocumentEmbeddingsByDocumentIDParams) ([]sqlc.CognitiveDocumentEmbedding, error) {
	return s.store.GetDocumentEmbeddingsByDocumentID(ctx, arg)
}

func (s *embeddingStore) SearchSimilarDocuments(ctx context.Context, arg sqlc.SearchSimilarDocumentsParams) ([]sqlc.SearchSimilarDocumentsRow, error) {
	return s.store.SearchSimilarDocuments(ctx, arg)
}

func (s *embeddingStore) DeleteDocumentEmbeddings(ctx context.Context, arg sqlc.DeleteDocumentEmbeddingsParams) error {
	return s.store.DeleteDocumentEmbeddings(ctx, arg)
}

func (s *embeddingStore) CountDocumentEmbeddingsByOrganization(ctx context.Context, organizationID int32) (int64, error) {
	return s.store.CountDocumentEmbeddingsByOrganization(ctx, organizationID)
}

// chatStore implements adapters.ChatStore
type chatStore struct {
	store sqlc.Store
}

func NewChatStore(store sqlc.Store) adapters.ChatStore {
	return &chatStore{store: store}
}

// Sessions

func (s *chatStore) CreateChatSession(ctx context.Context, arg sqlc.CreateChatSessionParams) (sqlc.CognitiveChatSession, error) {
	return s.store.CreateChatSession(ctx, arg)
}

func (s *chatStore) GetChatSessionByID(ctx context.Context, arg sqlc.GetChatSessionByIDParams) (sqlc.CognitiveChatSession, error) {
	return s.store.GetChatSessionByID(ctx, arg)
}

func (s *chatStore) ListChatSessionsByAccount(ctx context.Context, arg sqlc.ListChatSessionsByAccountParams) ([]sqlc.CognitiveChatSession, error) {
	return s.store.ListChatSessionsByAccount(ctx, arg)
}

func (s *chatStore) UpdateChatSessionTitle(ctx context.Context, arg sqlc.UpdateChatSessionTitleParams) (sqlc.CognitiveChatSession, error) {
	return s.store.UpdateChatSessionTitle(ctx, arg)
}

func (s *chatStore) DeleteChatSession(ctx context.Context, arg sqlc.DeleteChatSessionParams) error {
	return s.store.DeleteChatSession(ctx, arg)
}

// Messages

func (s *chatStore) CreateChatMessage(ctx context.Context, arg sqlc.CreateChatMessageParams) (sqlc.CognitiveChatMessage, error) {
	return s.store.CreateChatMessage(ctx, arg)
}

func (s *chatStore) GetChatMessagesBySession(ctx context.Context, sessionID int32) ([]sqlc.CognitiveChatMessage, error) {
	return s.store.GetChatMessagesBySession(ctx, sessionID)
}

func (s *chatStore) GetRecentChatMessages(ctx context.Context, arg sqlc.GetRecentChatMessagesParams) ([]sqlc.CognitiveChatMessage, error) {
	return s.store.GetRecentChatMessages(ctx, arg)
}

func (s *chatStore) CountChatMessagesBySession(ctx context.Context, sessionID int32) (int64, error) {
	return s.store.CountChatMessagesBySession(ctx, sessionID)
}

func (s *chatStore) DeleteChatMessage(ctx context.Context, id int32) error {
	return s.store.DeleteChatMessage(ctx, id)
}
