package repositories

import (
	"context"
	"fmt"

	"github.com/moasq/go-b2b-starter/app/example_cognitive/domain"
	"github.com/moasq/go-b2b-starter/pkg/db/adapters"
	sqlc "github.com/moasq/go-b2b-starter/pkg/db/postgres/sqlc/gen"
	"github.com/pgvector/pgvector-go"
)

type embeddingRepository struct {
	store adapters.EmbeddingStore
}

func NewEmbeddingRepository(store adapters.EmbeddingStore) domain.EmbeddingRepository {
	return &embeddingRepository{store: store}
}

func (r *embeddingRepository) Create(ctx context.Context, embedding *domain.DocumentEmbedding) (*domain.DocumentEmbedding, error) {
	params := sqlc.CreateDocumentEmbeddingParams{
		DocumentID:     embedding.DocumentID,
		OrganizationID: embedding.OrganizationID,
		Embedding:      toVector(embedding.Embedding),
		ContentHash:    toPgText(embedding.ContentHash),
		ContentPreview: toPgText(embedding.ContentPreview),
		ChunkIndex:     toPgInt4(embedding.ChunkIndex),
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
		Column1:        toVector(embedding),
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
				ContentHash:    fromPgText(result.ContentHash),
				ContentPreview: fromPgText(result.ContentPreview),
				ChunkIndex:     fromPgInt4(result.ChunkIndex),
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

// mapToDomain maps a database embedding to a domain embedding
func (r *embeddingRepository) mapToDomain(e *sqlc.CognitiveDocumentEmbedding) *domain.DocumentEmbedding {
	return &domain.DocumentEmbedding{
		ID:             e.ID,
		DocumentID:     e.DocumentID,
		OrganizationID: e.OrganizationID,
		Embedding:      fromVector(e.Embedding),
		ContentHash:    fromPgText(e.ContentHash),
		ContentPreview: fromPgText(e.ContentPreview),
		ChunkIndex:     fromPgInt4(e.ChunkIndex),
		CreatedAt:      e.CreatedAt.Time,
		UpdatedAt:      e.UpdatedAt.Time,
	}
}

// Vector conversion helpers
func toVector(embedding []float64) pgvector.Vector {
	floats := make([]float32, len(embedding))
	for i, v := range embedding {
		floats[i] = float32(v)
	}
	return pgvector.NewVector(floats)
}

func fromVector(v pgvector.Vector) []float64 {
	slice := v.Slice()
	floats := make([]float64, len(slice))
	for i, f := range slice {
		floats[i] = float64(f)
	}
	return floats
}
