package domain

import "context"

// DocumentRepository defines the interface for document data operations
type DocumentRepository interface {
	// Create creates a new document
	Create(ctx context.Context, doc *Document) (*Document, error)

	// GetByID retrieves a document by ID
	GetByID(ctx context.Context, orgID, docID int32) (*Document, error)

	// GetByFileAssetID retrieves a document by file asset ID
	GetByFileAssetID(ctx context.Context, orgID, fileAssetID int32) (*Document, error)

	// List retrieves documents with pagination
	List(ctx context.Context, orgID int32, limit, offset int32) ([]*Document, error)

	// ListByStatus retrieves documents by status with pagination
	ListByStatus(ctx context.Context, orgID int32, status DocumentStatus, limit, offset int32) ([]*Document, error)

	// UpdateStatus updates the document status
	UpdateStatus(ctx context.Context, orgID, docID int32, status DocumentStatus) (*Document, error)

	// UpdateExtractedText updates the extracted text and sets status to processed
	UpdateExtractedText(ctx context.Context, orgID, docID int32, text string) (*Document, error)

	// Update updates document metadata
	Update(ctx context.Context, doc *Document) (*Document, error)

	// Delete removes a document
	Delete(ctx context.Context, orgID, docID int32) error

	// Count returns the total count of documents for an organization
	Count(ctx context.Context, orgID int32) (int64, error)

	// CountByStatus returns the count of documents with a specific status
	CountByStatus(ctx context.Context, orgID int32, status DocumentStatus) (int64, error)
}
