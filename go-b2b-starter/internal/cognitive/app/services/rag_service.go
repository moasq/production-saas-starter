package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/moasq/go-b2b-starter/internal/cognitive/domain"
)

const (
	// DefaultMaxDocuments is the default number of documents to retrieve for RAG
	DefaultMaxDocuments = 3
	// DefaultContextHistory is the default number of messages to include in context
	DefaultContextHistory = 10
	// SystemPrompt is the default system prompt for RAG
	SystemPrompt = `You are a helpful assistant that answers questions based on the provided context.
If the context doesn't contain relevant information, say so clearly.
Always cite which documents you used to answer the question.`
)

type ragService struct {
	chatRepo          domain.ChatRepository
	embeddingRepo     domain.EmbeddingRepository
	textVectorizer    domain.TextVectorizer
	assistantProvider domain.AssistantProvider
}

func NewRAGService(
	chatRepo domain.ChatRepository,
	embeddingRepo domain.EmbeddingRepository,
	textVectorizer domain.TextVectorizer,
	assistantProvider domain.AssistantProvider,
) RAGService {
	return &ragService{
		chatRepo:          chatRepo,
		embeddingRepo:     embeddingRepo,
		textVectorizer:    textVectorizer,
		assistantProvider: assistantProvider,
	}
}

func (s *ragService) Chat(ctx context.Context, orgID, accountID int32, req *domain.ChatRequest) (*domain.ChatResponse, error) {
	var session *domain.ChatSession
	var err error

	// Get or create session
	if req.SessionID > 0 {
		session, err = s.chatRepo.GetSessionByID(ctx, orgID, req.SessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to get session: %w", err)
		}
	} else {
		// Create new session
		session = &domain.ChatSession{
			OrganizationID: orgID,
			AccountID:      accountID,
			Title:          generateSessionTitle(req.Message),
		}
		session, err = s.chatRepo.CreateSession(ctx, session)
		if err != nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}
	}

	// Save user message
	userMessage := &domain.ChatMessage{
		SessionID: session.ID,
		Role:      domain.ChatRoleUser,
		Content:   req.Message,
	}
	userMessage, err = s.chatRepo.CreateMessage(ctx, userMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// Build context and generate response
	var referencedDocs []*domain.SimilarDocument
	var prompt string

	if req.UseRAG {
		// Search for similar documents
		maxDocs := req.MaxDocuments
		if maxDocs <= 0 {
			maxDocs = DefaultMaxDocuments
		}

		// Generate embedding for the query and search
		embedding, err := s.textVectorizer.Vectorize(ctx, req.Message)
		if err == nil {
			docs, err := s.embeddingRepo.SearchSimilar(ctx, orgID, embedding, int32(maxDocs))
			if err == nil {
				referencedDocs = docs
			}
		}

		// Build RAG prompt
		prompt = s.buildRAGPrompt(req.Message, referencedDocs)
	} else {
		prompt = req.Message
	}

	// Get conversation history for context
	contextHistory := req.ContextHistory
	if contextHistory <= 0 {
		contextHistory = DefaultContextHistory
	}

	history, _ := s.chatRepo.GetRecentMessages(ctx, session.ID, int32(contextHistory))

	// Build full prompt with history
	fullPrompt := s.buildPromptWithHistory(prompt, history)

	// Generate response using AI assistant
	response, err := s.assistantProvider.GenerateResponse(ctx, fullPrompt)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrRAGCompletionFailed, err)
	}

	// Extract document IDs from referenced docs
	var docIDs []int32
	for _, doc := range referencedDocs {
		docIDs = append(docIDs, doc.DocumentID)
	}

	// Save assistant response
	assistantMessage := &domain.ChatMessage{
		SessionID:      session.ID,
		Role:           domain.ChatRoleAssistant,
		Content:        response.Content,
		ReferencedDocs: docIDs,
		TokensUsed:     int32(response.TokensUsed),
	}
	assistantMessage, err = s.chatRepo.CreateMessage(ctx, assistantMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to save assistant message: %w", err)
	}

	// Convert []*SimilarDocument to []SimilarDocument
	var docs []domain.SimilarDocument
	for _, doc := range referencedDocs {
		if doc != nil {
			docs = append(docs, *doc)
		}
	}

	return &domain.ChatResponse{
		SessionID:      session.ID,
		Message:        assistantMessage,
		ReferencedDocs: docs,
		TokensUsed:     int32(response.TokensUsed),
	}, nil
}

func (s *ragService) GetSession(ctx context.Context, orgID, sessionID int32) (*domain.ChatSession, error) {
	return s.chatRepo.GetSessionByID(ctx, orgID, sessionID)
}

func (s *ragService) ListSessions(ctx context.Context, orgID, accountID int32, limit, offset int32) ([]*domain.ChatSession, error) {
	return s.chatRepo.ListSessionsByAccount(ctx, orgID, accountID, limit, offset)
}

func (s *ragService) DeleteSession(ctx context.Context, orgID, sessionID int32) error {
	return s.chatRepo.DeleteSession(ctx, orgID, sessionID)
}

func (s *ragService) GetSessionHistory(ctx context.Context, orgID, sessionID int32) ([]*domain.ChatMessage, error) {
	// Verify session belongs to organization
	_, err := s.chatRepo.GetSessionByID(ctx, orgID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify session: %w", err)
	}

	return s.chatRepo.GetMessagesBySession(ctx, sessionID)
}

func (s *ragService) UpdateSessionTitle(ctx context.Context, orgID, sessionID int32, title string) (*domain.ChatSession, error) {
	return s.chatRepo.UpdateSessionTitle(ctx, orgID, sessionID, title)
}

// buildRAGPrompt builds a prompt with RAG context
func (s *ragService) buildRAGPrompt(query string, docs []*domain.SimilarDocument) string {
	if len(docs) == 0 {
		return fmt.Sprintf("%s\n\nUser Question: %s", SystemPrompt, query)
	}

	var contextBuilder strings.Builder
	contextBuilder.WriteString(SystemPrompt)
	contextBuilder.WriteString("\n\n--- CONTEXT FROM DOCUMENTS ---\n")

	for i, doc := range docs {
		contextBuilder.WriteString(fmt.Sprintf("\n[Document %d (similarity: %.2f)]:\n%s\n",
			i+1, doc.SimilarityScore, doc.ContentPreview))
	}

	contextBuilder.WriteString("\n--- END OF CONTEXT ---\n\n")
	contextBuilder.WriteString(fmt.Sprintf("User Question: %s", query))

	return contextBuilder.String()
}

// buildPromptWithHistory builds a prompt including conversation history
func (s *ragService) buildPromptWithHistory(prompt string, history []*domain.ChatMessage) string {
	if len(history) == 0 {
		return prompt
	}

	var builder strings.Builder
	builder.WriteString("Previous conversation:\n")

	// History is in descending order, so reverse it
	for i := len(history) - 1; i >= 0; i-- {
		msg := history[i]
		role := "User"
		if msg.Role == domain.ChatRoleAssistant {
			role = "Assistant"
		}
		builder.WriteString(fmt.Sprintf("%s: %s\n", role, msg.Content))
	}

	builder.WriteString("\nCurrent prompt:\n")
	builder.WriteString(prompt)

	return builder.String()
}

// generateSessionTitle generates a title from the first message
func generateSessionTitle(message string) string {
	// Take first 50 characters of the message as title
	if len(message) <= 50 {
		return message
	}
	return message[:50] + "..."
}
