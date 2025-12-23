package repositories

import (
	"context"
	"fmt"

	"github.com/moasq/go-b2b-starter/internal/db/helpers"
	sqlc "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
	"github.com/moasq/go-b2b-starter/internal/modules/cognitive/domain"
)

// embeddingRepository implements domain.EmbeddingRepository using SQLC internally.
// SQLC types are never exposed outside this package.
type embeddingRepository struct {
	store sqlc.Store
}

// NewEmbeddingRepository creates a new EmbeddingRepository implementation.
func NewEmbeddingRepository(store sqlc.Store) domain.EmbeddingRepository {
	return &embeddingRepository{store: store}
}

func (r *embeddingRepository) Create(ctx context.Context, embedding *domain.DocumentEmbedding) (*domain.DocumentEmbedding, error) {
	params := sqlc.CreateDocumentEmbeddingParams{
		DocumentID:     embedding.DocumentID,
		OrganizationID: embedding.OrganizationID,
		Embedding:      helpers.ToVector(embedding.Embedding),
		ContentHash:    helpers.ToPgText(embedding.ContentHash),
		ContentPreview: helpers.ToPgText(embedding.ContentPreview),
		ChunkIndex:     helpers.ToPgInt4(embedding.ChunkIndex),
	}

	result, err := r.store.CreateDocumentEmbedding(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create document embedding: %w", err)
	}

	return r.mapToDomain(&result), nil
}

func (r *embeddingRepository) GetByID(ctx context.Context, orgID, embeddingID int32) (*domain.DocumentEmbedding, error) {
	params := sqlc.GetDocumentEmbeddingByIDParams{
		ID:             embeddingID,
		OrganizationID: orgID,
	}

	result, err := r.store.GetDocumentEmbeddingByID(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get document embedding: %w", err)
	}

	return r.mapToDomain(&result), nil
}

func (r *embeddingRepository) GetByDocumentID(ctx context.Context, orgID, documentID int32) ([]*domain.DocumentEmbedding, error) {
	params := sqlc.GetDocumentEmbeddingsByDocumentIDParams{
		DocumentID:     documentID,
		OrganizationID: orgID,
	}

	results, err := r.store.GetDocumentEmbeddingsByDocumentID(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get document embeddings: %w", err)
	}

	embeddings := make([]*domain.DocumentEmbedding, len(results))
	for i, result := range results {
		embeddings[i] = r.mapToDomain(&result)
	}

	return embeddings, nil
}

func (r *embeddingRepository) SearchSimilar(ctx context.Context, orgID int32, embedding []float64, limit int32) ([]*domain.SimilarDocument, error) {
	params := sqlc.SearchSimilarDocumentsParams{
		Column1:        helpers.ToVector(embedding),
		OrganizationID: orgID,
		Limit:          limit,
	}

	results, err := r.store.SearchSimilarDocuments(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar documents: %w", err)
	}

	docs := make([]*domain.SimilarDocument, len(results))
	for i, result := range results {
		docs[i] = &domain.SimilarDocument{
			DocumentEmbedding: domain.DocumentEmbedding{
				ID:             result.ID,
				DocumentID:     result.DocumentID,
				OrganizationID: result.OrganizationID,
				ContentHash:    helpers.FromPgText(result.ContentHash),
				ContentPreview: helpers.FromPgText(result.ContentPreview),
				ChunkIndex:     helpers.FromPgInt4(result.ChunkIndex),
				CreatedAt:      result.CreatedAt.Time,
				UpdatedAt:      result.UpdatedAt.Time,
			},
			SimilarityScore: result.SimilarityScore,
		}
	}

	return docs, nil
}

func (r *embeddingRepository) Delete(ctx context.Context, orgID, documentID int32) error {
	params := sqlc.DeleteDocumentEmbeddingsParams{
		DocumentID:     documentID,
		OrganizationID: orgID,
	}

	if err := r.store.DeleteDocumentEmbeddings(ctx, params); err != nil {
		return fmt.Errorf("failed to delete document embeddings: %w", err)
	}

	return nil
}

func (r *embeddingRepository) Count(ctx context.Context, orgID int32) (int64, error) {
	count, err := r.store.CountDocumentEmbeddingsByOrganization(ctx, orgID)
	if err != nil {
		return 0, fmt.Errorf("failed to count document embeddings: %w", err)
	}

	return count, nil
}

// mapToDomain maps SQLC embedding type to domain type.
// This is the translation boundary - SQLC types never escape this function.
func (r *embeddingRepository) mapToDomain(e *sqlc.CognitiveDocumentEmbedding) *domain.DocumentEmbedding {
	return &domain.DocumentEmbedding{
		ID:             e.ID,
		DocumentID:     e.DocumentID,
		OrganizationID: e.OrganizationID,
		Embedding:      helpers.FromVector(e.Embedding),
		ContentHash:    helpers.FromPgText(e.ContentHash),
		ContentPreview: helpers.FromPgText(e.ContentPreview),
		ChunkIndex:     helpers.FromPgInt4(e.ChunkIndex),
		CreatedAt:      e.CreatedAt.Time,
		UpdatedAt:      e.UpdatedAt.Time,
	}
}
