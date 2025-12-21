package adapters

import (
	"context"

	db "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
)

// DocumentStore provides database operations for documents
type DocumentStore interface {
	CreateDocument(ctx context.Context, arg db.CreateDocumentParams) (db.DocumentsDocument, error)
	GetDocumentByID(ctx context.Context, arg db.GetDocumentByIDParams) (db.DocumentsDocument, error)
	GetDocumentByFileAssetID(ctx context.Context, arg db.GetDocumentByFileAssetIDParams) (db.DocumentsDocument, error)
	ListDocumentsByOrganization(ctx context.Context, arg db.ListDocumentsByOrganizationParams) ([]db.DocumentsDocument, error)
	ListDocumentsByStatus(ctx context.Context, arg db.ListDocumentsByStatusParams) ([]db.DocumentsDocument, error)
	UpdateDocumentStatus(ctx context.Context, arg db.UpdateDocumentStatusParams) (db.DocumentsDocument, error)
	UpdateDocumentExtractedText(ctx context.Context, arg db.UpdateDocumentExtractedTextParams) (db.DocumentsDocument, error)
	UpdateDocument(ctx context.Context, arg db.UpdateDocumentParams) (db.DocumentsDocument, error)
	DeleteDocument(ctx context.Context, arg db.DeleteDocumentParams) error
	CountDocumentsByOrganization(ctx context.Context, organizationID int32) (int64, error)
	CountDocumentsByStatus(ctx context.Context, arg db.CountDocumentsByStatusParams) (int64, error)
}
