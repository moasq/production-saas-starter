package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/moasq/go-b2b-starter/internal/modules/cognitive/domain"
)

const (
	// MaxChunkSize is the maximum number of characters per chunk
	MaxChunkSize = 8000
	// ContentPreviewLength is the length of content preview to store
	ContentPreviewLength = 500
)

type embeddingService struct {
	embeddingRepo  domain.EmbeddingRepository
	textVectorizer domain.TextVectorizer
}

func NewEmbeddingService(
	embeddingRepo domain.EmbeddingRepository,
	textVectorizer domain.TextVectorizer,
) EmbeddingService {
	return &embeddingService{
		embeddingRepo:  embeddingRepo,
		textVectorizer: textVectorizer,
	}
}

func (s *embeddingService) EmbedDocument(ctx context.Context, orgID, documentID int32, text string) (*domain.DocumentEmbedding, error) {
	// Generate embedding using text vectorizer
	embedding, err := s.textVectorizer.Vectorize(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrEmbeddingGenerationFailed, err)
	}

	// Create content hash for deduplication
	contentHash := s.hashContent(text)

	// Create content preview
	contentPreview := text
	if len(contentPreview) > ContentPreviewLength {
		contentPreview = contentPreview[:ContentPreviewLength]
	}

	// Create embedding record
	docEmbedding := &domain.DocumentEmbedding{
		DocumentID:     documentID,
		OrganizationID: orgID,
		Embedding:      embedding,
		ContentHash:    contentHash,
		ContentPreview: contentPreview,
		ChunkIndex:     0, // Single chunk for now
	}

	result, err := s.embeddingRepo.Create(ctx, docEmbedding)
	if err != nil {
		return nil, fmt.Errorf("failed to store embedding: %w", err)
	}

	return result, nil
}

func (s *embeddingService) GetDocumentEmbeddings(ctx context.Context, orgID, documentID int32) ([]*domain.DocumentEmbedding, error) {
	return s.embeddingRepo.GetByDocumentID(ctx, orgID, documentID)
}

func (s *embeddingService) SearchSimilarDocuments(ctx context.Context, orgID int32, text string, limit int32) ([]*domain.SimilarDocument, error) {
	// Generate embedding for the search query
	embedding, err := s.textVectorizer.Vectorize(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrEmbeddingGenerationFailed, err)
	}

	// Search for similar documents
	return s.embeddingRepo.SearchSimilar(ctx, orgID, embedding, limit)
}

func (s *embeddingService) DeleteDocumentEmbeddings(ctx context.Context, orgID, documentID int32) error {
	if err := s.embeddingRepo.Delete(ctx, orgID, documentID); err != nil {
		return fmt.Errorf("failed to delete embeddings: %w", err)
	}

	return nil
}

func (s *embeddingService) GetStats(ctx context.Context, orgID int32) (*domain.EmbeddingStats, error) {
	count, err := s.embeddingRepo.Count(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding count: %w", err)
	}

	return &domain.EmbeddingStats{
		TotalEmbeddings: count,
		TotalDocuments:  count, // For now, 1:1 relationship
	}, nil
}

// hashContent creates a SHA-256 hash of the content for deduplication
func (s *embeddingService) hashContent(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}
