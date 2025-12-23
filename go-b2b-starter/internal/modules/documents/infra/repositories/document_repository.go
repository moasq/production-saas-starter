package repositories

import (
	"context"
	"fmt"

	"github.com/moasq/go-b2b-starter/internal/db/helpers"
	sqlc "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
	"github.com/moasq/go-b2b-starter/internal/modules/documents/domain"
)

// documentRepository implements domain.DocumentRepository using SQLC internally.
// SQLC types are never exposed outside this package.
type documentRepository struct {
	store sqlc.Store
}

// NewDocumentRepository creates a new DocumentRepository implementation.
func NewDocumentRepository(store sqlc.Store) domain.DocumentRepository {
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
		ExtractedText:  helpers.ToPgText(doc.ExtractedText),
		Status:         string(doc.Status),
		Metadata:       helpers.ToJSONB(doc.Metadata),
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
		ExtractedText:  helpers.ToPgText(text),
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
		Metadata:       helpers.ToJSONB(doc.Metadata),
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

// mapToDomain converts SQLC document type to domain type.
// This is the translation boundary - SQLC types never escape this function.
func (r *documentRepository) mapToDomain(doc *sqlc.DocumentsDocument) *domain.Document {
	return &domain.Document{
		ID:             doc.ID,
		OrganizationID: doc.OrganizationID,
		FileAssetID:    doc.FileAssetID,
		Title:          doc.Title,
		FileName:       doc.FileName,
		ContentType:    doc.ContentType,
		FileSize:       doc.FileSize,
		ExtractedText:  helpers.FromPgText(doc.ExtractedText),
		Status:         domain.DocumentStatus(doc.Status),
		Metadata:       helpers.FromJSONB(doc.Metadata),
		CreatedAt:      doc.CreatedAt.Time,
		UpdatedAt:      doc.UpdatedAt.Time,
	}
}
