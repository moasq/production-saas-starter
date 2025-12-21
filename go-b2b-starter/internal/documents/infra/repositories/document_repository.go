package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/moasq/go-b2b-starter/internal/documents/domain"
	"github.com/moasq/go-b2b-starter/internal/db/adapters"
	sqlc "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
)

type documentRepository struct {
	store adapters.DocumentStore
}

func NewDocumentRepository(store adapters.DocumentStore) domain.DocumentRepository {
	return &documentRepository{store: store}
}

func (r *documentRepository) Create(ctx context.Context, doc *domain.Document) (*domain.Document, error) {
	params := sqlc.CreateDocumentParams{
		OrganizationID: doc.OrganizationID,
		FileAssetID:    doc.FileAssetID,
		Title:          doc.Title,
		FileName:       doc.FileName,
		ContentType:    doc.ContentType,
		FileSize:       doc.FileSize,
		ExtractedText:  toPgText(doc.ExtractedText),
		Status:         string(doc.Status),
		Metadata:       toJSONB(doc.Metadata),
	}

	result, err := r.store.CreateDocument(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	return r.mapToDomain(&result), nil
}

func (r *documentRepository) GetByID(ctx context.Context, orgID, docID int32) (*domain.Document, error) {
	params := sqlc.GetDocumentByIDParams{
		ID:             docID,
		OrganizationID: orgID,
	}

	result, err := r.store.GetDocumentByID(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return r.mapToDomain(&result), nil
}

func (r *documentRepository) GetByFileAssetID(ctx context.Context, orgID, fileAssetID int32) (*domain.Document, error) {
	params := sqlc.GetDocumentByFileAssetIDParams{
		FileAssetID:    fileAssetID,
		OrganizationID: orgID,
	}

	result, err := r.store.GetDocumentByFileAssetID(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get document by file asset: %w", err)
	}

	return r.mapToDomain(&result), nil
}

func (r *documentRepository) List(ctx context.Context, orgID int32, limit, offset int32) ([]*domain.Document, error) {
	params := sqlc.ListDocumentsByOrganizationParams{
		OrganizationID: orgID,
		Limit:          limit,
		Offset:         offset,
	}

	results, err := r.store.ListDocumentsByOrganization(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	docs := make([]*domain.Document, len(results))
	for i, result := range results {
		docs[i] = r.mapToDomain(&result)
	}

	return docs, nil
}

func (r *documentRepository) ListByStatus(ctx context.Context, orgID int32, status domain.DocumentStatus, limit, offset int32) ([]*domain.Document, error) {
	params := sqlc.ListDocumentsByStatusParams{
		OrganizationID: orgID,
		Status:         string(status),
		Limit:          limit,
		Offset:         offset,
	}

	results, err := r.store.ListDocumentsByStatus(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents by status: %w", err)
	}

	docs := make([]*domain.Document, len(results))
	for i, result := range results {
		docs[i] = r.mapToDomain(&result)
	}

	return docs, nil
}

func (r *documentRepository) UpdateStatus(ctx context.Context, orgID, docID int32, status domain.DocumentStatus) (*domain.Document, error) {
	params := sqlc.UpdateDocumentStatusParams{
		ID:             docID,
		OrganizationID: orgID,
		Status:         string(status),
	}

	result, err := r.store.UpdateDocumentStatus(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update document status: %w", err)
	}

	return r.mapToDomain(&result), nil
}

func (r *documentRepository) UpdateExtractedText(ctx context.Context, orgID, docID int32, text string) (*domain.Document, error) {
	params := sqlc.UpdateDocumentExtractedTextParams{
		ID:             docID,
		OrganizationID: orgID,
		ExtractedText:  toPgText(text),
	}

	result, err := r.store.UpdateDocumentExtractedText(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update extracted text: %w", err)
	}

	return r.mapToDomain(&result), nil
}

func (r *documentRepository) Update(ctx context.Context, doc *domain.Document) (*domain.Document, error) {
	params := sqlc.UpdateDocumentParams{
		ID:             doc.ID,
		OrganizationID: doc.OrganizationID,
		Title:          doc.Title,
		Metadata:       toJSONB(doc.Metadata),
	}

	result, err := r.store.UpdateDocument(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	return r.mapToDomain(&result), nil
}

func (r *documentRepository) Delete(ctx context.Context, orgID, docID int32) error {
	params := sqlc.DeleteDocumentParams{
		ID:             docID,
		OrganizationID: orgID,
	}

	if err := r.store.DeleteDocument(ctx, params); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

func (r *documentRepository) Count(ctx context.Context, orgID int32) (int64, error) {
	count, err := r.store.CountDocumentsByOrganization(ctx, orgID)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	return count, nil
}

func (r *documentRepository) CountByStatus(ctx context.Context, orgID int32, status domain.DocumentStatus) (int64, error) {
	params := sqlc.CountDocumentsByStatusParams{
		OrganizationID: orgID,
		Status:         string(status),
	}

	count, err := r.store.CountDocumentsByStatus(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents by status: %w", err)
	}

	return count, nil
}

// mapToDomain maps a database document to a domain document
func (r *documentRepository) mapToDomain(doc *sqlc.DocumentsDocument) *domain.Document {
	return &domain.Document{
		ID:             doc.ID,
		OrganizationID: doc.OrganizationID,
		FileAssetID:    doc.FileAssetID,
		Title:          doc.Title,
		FileName:       doc.FileName,
		ContentType:    doc.ContentType,
		FileSize:       doc.FileSize,
		ExtractedText:  fromPgText(doc.ExtractedText),
		Status:         domain.DocumentStatus(doc.Status),
		Metadata:       fromJSONB(doc.Metadata),
		CreatedAt:      doc.CreatedAt.Time,
		UpdatedAt:      doc.UpdatedAt.Time,
	}
}

// Helper functions for type conversion
func toPgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

func fromPgText(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

func toJSONB(m map[string]interface{}) []byte {
	if m == nil {
		return []byte("{}")
	}
	// In production, use proper JSON marshaling
	return []byte("{}")
}

func fromJSONB(b []byte) map[string]interface{} {
	if b == nil {
		return nil
	}
	// In production, use proper JSON unmarshaling
	return nil
}
