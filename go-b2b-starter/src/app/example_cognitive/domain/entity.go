package domain

import (
	"time"
)

// ChatRole represents the role of a message sender
type ChatRole string

const (
	ChatRoleUser      ChatRole = "user"
	ChatRoleAssistant ChatRole = "assistant"
	ChatRoleSystem    ChatRole = "system"
)

// DocumentEmbedding represents a vector embedding for a document
type DocumentEmbedding struct {
	ID             int32     `json:"id"`
	DocumentID     int32     `json:"document_id"`
	OrganizationID int32     `json:"organization_id"`
	Embedding      []float64 `json:"embedding,omitempty"` // 1536 dimensions for OpenAI
	ContentHash    string    `json:"content_hash,omitempty"`
	ContentPreview string    `json:"content_preview,omitempty"`
	ChunkIndex     int32     `json:"chunk_index"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// SimilarDocument represents a document found through similarity search
type SimilarDocument struct {
	DocumentEmbedding
	SimilarityScore float64 `json:"similarity_score"`
}

// ChatSession represents a conversation session
type ChatSession struct {
	ID             int32     `json:"id"`
	OrganizationID int32     `json:"organization_id"`
	AccountID      int32     `json:"account_id"`
	Title          string    `json:"title,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (s *ChatSession) GetID() int32 {
	return s.ID
}

// Validate validates the chat session entity
func (s *ChatSession) Validate() error {
	if s.OrganizationID == 0 {
		return ErrSessionOrganizationRequired
	}
	if s.AccountID == 0 {
		return ErrSessionAccountRequired
	}
	return nil
}

// ChatMessage represents a message within a chat session
type ChatMessage struct {
	ID             int32     `json:"id"`
	SessionID      int32     `json:"session_id"`
	Role           ChatRole  `json:"role"`
	Content        string    `json:"content"`
	ReferencedDocs []int32   `json:"referenced_docs,omitempty"`
	TokensUsed     int32     `json:"tokens_used,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

func (m *ChatMessage) GetID() int32 {
	return m.ID
}

// Validate validates the chat message entity
func (m *ChatMessage) Validate() error {
	if m.SessionID == 0 {
		return ErrMessageSessionRequired
	}
	if m.Content == "" {
		return ErrMessageContentRequired
	}
	if m.Role == "" {
		return ErrMessageRoleRequired
	}
	return nil
}

func (m *ChatMessage) IsUserMessage() bool {
	return m.Role == ChatRoleUser
}

func (m *ChatMessage) IsAssistantMessage() bool {
	return m.Role == ChatRoleAssistant
}

// RAGContext represents context retrieved for RAG
type RAGContext struct {
	Documents []SimilarDocument `json:"documents"`
	Query     string            `json:"query"`
}

// ChatRequest represents a request to send a chat message
type ChatRequest struct {
	SessionID      int32  `json:"session_id,omitempty"` // Optional - create new session if not provided
	Message        string `json:"message"`
	UseRAG         bool   `json:"use_rag,omitempty"` // Whether to use RAG for context
	MaxDocuments   int    `json:"max_documents,omitempty"`
	ContextHistory int    `json:"context_history,omitempty"` // Number of previous messages to include
}

// ChatResponse represents a response from the chat service
type ChatResponse struct {
	SessionID        int32             `json:"session_id"`
	Message          *ChatMessage      `json:"message"`
	ReferencedDocs   []SimilarDocument `json:"referenced_docs,omitempty"`
	TokensUsed       int32             `json:"tokens_used,omitempty"`
}

// EmbeddingStats represents embedding statistics
type EmbeddingStats struct {
	TotalEmbeddings int64 `json:"total_embeddings"`
	TotalDocuments  int64 `json:"total_documents"`
}

// ChatStats represents chat statistics
type ChatStats struct {
	TotalSessions int64 `json:"total_sessions"`
	TotalMessages int64 `json:"total_messages"`
}
