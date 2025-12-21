package services

import (
	"context"
	"io"

	"github.com/moasq/go-b2b-starter/internal/documents/domain"
)

// DocumentService defines the interface for document operations
type DocumentService interface {
	// UploadDocument uploads a new document and extracts text from it
	UploadDocument(ctx context.Context, orgID int32, req *UploadDocumentRequest, content io.Reader) (*domain.Document, error)

	// GetDocument retrieves a document by ID
	GetDocument(ctx context.Context, orgID, docID int32) (*domain.Document, error)

	// ListDocuments lists documents with pagination
	ListDocuments(ctx context.Context, orgID int32, req *ListDocumentsRequest) (*ListDocumentsResponse, error)

	// UpdateDocument updates document metadata
	UpdateDocument(ctx context.Context, orgID, docID int32, req *UpdateDocumentRequest) (*domain.Document, error)

	// DeleteDocument deletes a document
	DeleteDocument(ctx context.Context, orgID, docID int32) error

	// GetDocumentStats retrieves document statistics
	GetDocumentStats(ctx context.Context, orgID int32) (*domain.DocumentStats, error)

	// ProcessDocument processes a document (extract text, etc.)
	ProcessDocument(ctx context.Context, orgID, docID int32) (*domain.Document, error)
}

// UploadDocumentRequest represents a request to upload a document
type UploadDocumentRequest struct {
	Title       string                 `json:"title"`
	FileName    string                 `json:"file_name"`
	ContentType string                 `json:"content_type"`
	FileSize    int64                  `json:"file_size"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ListDocumentsRequest represents a request to list documents
type ListDocumentsRequest struct {
	Status *domain.DocumentStatus `json:"status,omitempty"`
	Limit  int32                  `json:"limit"`
	Offset int32                  `json:"offset"`
}

// ListDocumentsResponse represents the response for listing documents
type ListDocumentsResponse struct {
	Documents []*domain.Document `json:"documents"`
	Total     int64              `json:"total"`
	Limit     int32              `json:"limit"`
	Offset    int32              `json:"offset"`
}

// UpdateDocumentRequest represents a request to update a document
type UpdateDocumentRequest struct {
	Title    string                 `json:"title,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
