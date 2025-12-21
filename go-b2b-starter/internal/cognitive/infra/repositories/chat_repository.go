package repositories

import (
	"context"
	"fmt"

	"github.com/moasq/go-b2b-starter/internal/cognitive/domain"
	"github.com/moasq/go-b2b-starter/internal/db/adapters"
	sqlc "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
)

type chatRepository struct {
	store adapters.ChatStore
}

func NewChatRepository(store adapters.ChatStore) domain.ChatRepository {
	return &chatRepository{store: store}
}

// Sessions

func (r *chatRepository) CreateSession(ctx context.Context, session *domain.ChatSession) (*domain.ChatSession, error) {
	params := sqlc.CreateChatSessionParams{
		OrganizationID: session.OrganizationID,
		AccountID:      session.AccountID,
		Title:          toPgText(session.Title),
	}

	result, err := r.store.CreateChatSession(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat session: %w", err)
	}

	return r.mapSessionToDomain(&result), nil
}

func (r *chatRepository) GetSessionByID(ctx context.Context, orgID, sessionID int32) (*domain.ChatSession, error) {
	params := sqlc.GetChatSessionByIDParams{
		ID:             sessionID,
		OrganizationID: orgID,
	}

	result, err := r.store.GetChatSessionByID(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat session: %w", err)
	}

	return r.mapSessionToDomain(&result), nil
}

func (r *chatRepository) ListSessionsByAccount(ctx context.Context, orgID, accountID int32, limit, offset int32) ([]*domain.ChatSession, error) {
	params := sqlc.ListChatSessionsByAccountParams{
		OrganizationID: orgID,
		AccountID:      accountID,
		Limit:          limit,
		Offset:         offset,
	}

	results, err := r.store.ListChatSessionsByAccount(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list chat sessions: %w", err)
	}

	sessions := make([]*domain.ChatSession, len(results))
	for i, result := range results {
		sessions[i] = r.mapSessionToDomain(&result)
	}

	return sessions, nil
}

func (r *chatRepository) UpdateSessionTitle(ctx context.Context, orgID, sessionID int32, title string) (*domain.ChatSession, error) {
	params := sqlc.UpdateChatSessionTitleParams{
		ID:             sessionID,
		OrganizationID: orgID,
		Title:          toPgText(title),
	}

	result, err := r.store.UpdateChatSessionTitle(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update chat session title: %w", err)
	}

	return r.mapSessionToDomain(&result), nil
}

func (r *chatRepository) DeleteSession(ctx context.Context, orgID, sessionID int32) error {
	params := sqlc.DeleteChatSessionParams{
		ID:             sessionID,
		OrganizationID: orgID,
	}

	if err := r.store.DeleteChatSession(ctx, params); err != nil {
		return fmt.Errorf("failed to delete chat session: %w", err)
	}

	return nil
}

// Messages

func (r *chatRepository) CreateMessage(ctx context.Context, message *domain.ChatMessage) (*domain.ChatMessage, error) {
	params := sqlc.CreateChatMessageParams{
		SessionID:      message.SessionID,
		Role:           string(message.Role),
		Content:        message.Content,
		ReferencedDocs: message.ReferencedDocs,
		TokensUsed:     toPgInt4(message.TokensUsed),
	}

	result, err := r.store.CreateChatMessage(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat message: %w", err)
	}

	return r.mapMessageToDomain(&result), nil
}

func (r *chatRepository) GetMessagesBySession(ctx context.Context, sessionID int32) ([]*domain.ChatMessage, error) {
	results, err := r.store.GetChatMessagesBySession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat messages: %w", err)
	}

	messages := make([]*domain.ChatMessage, len(results))
	for i, result := range results {
		messages[i] = r.mapMessageToDomain(&result)
	}

	return messages, nil
}

func (r *chatRepository) GetRecentMessages(ctx context.Context, sessionID int32, limit int32) ([]*domain.ChatMessage, error) {
	params := sqlc.GetRecentChatMessagesParams{
		SessionID: sessionID,
		Limit:     limit,
	}

	results, err := r.store.GetRecentChatMessages(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent chat messages: %w", err)
	}

	messages := make([]*domain.ChatMessage, len(results))
	for i, result := range results {
		messages[i] = r.mapMessageToDomain(&result)
	}

	return messages, nil
}

func (r *chatRepository) CountMessagesBySession(ctx context.Context, sessionID int32) (int64, error) {
	count, err := r.store.CountChatMessagesBySession(ctx, sessionID)
	if err != nil {
		return 0, fmt.Errorf("failed to count chat messages: %w", err)
	}

	return count, nil
}

func (r *chatRepository) DeleteMessage(ctx context.Context, messageID int32) error {
	if err := r.store.DeleteChatMessage(ctx, messageID); err != nil {
		return fmt.Errorf("failed to delete chat message: %w", err)
	}

	return nil
}

// mapSessionToDomain maps a database session to a domain session
func (r *chatRepository) mapSessionToDomain(s *sqlc.CognitiveChatSession) *domain.ChatSession {
	return &domain.ChatSession{
		ID:             s.ID,
		OrganizationID: s.OrganizationID,
		AccountID:      s.AccountID,
		Title:          fromPgText(s.Title),
		CreatedAt:      s.CreatedAt.Time,
		UpdatedAt:      s.UpdatedAt.Time,
	}
}

// mapMessageToDomain maps a database message to a domain message
func (r *chatRepository) mapMessageToDomain(m *sqlc.CognitiveChatMessage) *domain.ChatMessage {
	return &domain.ChatMessage{
		ID:             m.ID,
		SessionID:      m.SessionID,
		Role:           domain.ChatRole(m.Role),
		Content:        m.Content,
		ReferencedDocs: m.ReferencedDocs,
		TokensUsed:     fromPgInt4(m.TokensUsed),
		CreatedAt:      m.CreatedAt.Time,
	}
}
