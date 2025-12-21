package adapterimpl

import (
	"context"

	"github.com/moasq/go-b2b-starter/internal/db/adapters"
	sqlc "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
)

// documentStore implements adapters.DocumentStore
type documentStore struct {
	store sqlc.Store
}

func NewDocumentStore(store sqlc.Store) adapters.DocumentStore {
	return &documentStore{store: store}
}

func (s *documentStore) CreateDocument(ctx context.Context, arg sqlc.CreateDocumentParams) (sqlc.DocumentsDocument, error) {
	return s.store.CreateDocument(ctx, arg)
}

func (s *documentStore) GetDocumentByID(ctx context.Context, arg sqlc.GetDocumentByIDParams) (sqlc.DocumentsDocument, error) {
	return s.store.GetDocumentByID(ctx, arg)
}

func (s *documentStore) GetDocumentByFileAssetID(ctx context.Context, arg sqlc.GetDocumentByFileAssetIDParams) (sqlc.DocumentsDocument, error) {
	return s.store.GetDocumentByFileAssetID(ctx, arg)
}

func (s *documentStore) ListDocumentsByOrganization(ctx context.Context, arg sqlc.ListDocumentsByOrganizationParams) ([]sqlc.DocumentsDocument, error) {
	return s.store.ListDocumentsByOrganization(ctx, arg)
}

func (s *documentStore) ListDocumentsByStatus(ctx context.Context, arg sqlc.ListDocumentsByStatusParams) ([]sqlc.DocumentsDocument, error) {
	return s.store.ListDocumentsByStatus(ctx, arg)
}

func (s *documentStore) UpdateDocumentStatus(ctx context.Context, arg sqlc.UpdateDocumentStatusParams) (sqlc.DocumentsDocument, error) {
	return s.store.UpdateDocumentStatus(ctx, arg)
}

func (s *documentStore) UpdateDocumentExtractedText(ctx context.Context, arg sqlc.UpdateDocumentExtractedTextParams) (sqlc.DocumentsDocument, error) {
	return s.store.UpdateDocumentExtractedText(ctx, arg)
}

func (s *documentStore) UpdateDocument(ctx context.Context, arg sqlc.UpdateDocumentParams) (sqlc.DocumentsDocument, error) {
	return s.store.UpdateDocument(ctx, arg)
}

func (s *documentStore) DeleteDocument(ctx context.Context, arg sqlc.DeleteDocumentParams) error {
	return s.store.DeleteDocument(ctx, arg)
}

func (s *documentStore) CountDocumentsByOrganization(ctx context.Context, organizationID int32) (int64, error) {
	return s.store.CountDocumentsByOrganization(ctx, organizationID)
}

func (s *documentStore) CountDocumentsByStatus(ctx context.Context, arg sqlc.CountDocumentsByStatusParams) (int64, error) {
	return s.store.CountDocumentsByStatus(ctx, arg)
}
