package domain

import (
	"time"
)

// DocumentStatus represents the processing status of a document
type DocumentStatus string

const (
	DocumentStatusPending    DocumentStatus = "pending"
	DocumentStatusProcessing DocumentStatus = "processing"
	DocumentStatusProcessed  DocumentStatus = "processed"
	DocumentStatusFailed     DocumentStatus = "failed"
)

// Document represents an uploaded document (PDF)
type Document struct {
	ID             int32                  `json:"id"`
	OrganizationID int32                  `json:"organization_id"`
	FileAssetID    int32                  `json:"file_asset_id"`
	Title          string                 `json:"title"`
	FileName       string                 `json:"file_name"`
	ContentType    string                 `json:"content_type"`
	FileSize       int64                  `json:"file_size"`
	ExtractedText  string                 `json:"extracted_text,omitempty"`
	Status         DocumentStatus         `json:"status"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

func (d *Document) GetID() int32 {
	return d.ID
}

// Validate validates the document entity
func (d *Document) Validate() error {
	if d.OrganizationID == 0 {
		return ErrDocumentOrganizationRequired
	}
	if d.Title == "" {
		return ErrDocumentTitleRequired
	}
	if d.FileName == "" {
		return ErrDocumentFileNameRequired
	}
	if d.FileAssetID == 0 {
		return ErrDocumentFileAssetRequired
	}
	return nil
}

func (d *Document) IsProcessed() bool {
	return d.Status == DocumentStatusProcessed
}

func (d *Document) IsPending() bool {
	return d.Status == DocumentStatusPending
}

func (d *Document) HasText() bool {
	return d.ExtractedText != ""
}

// DocumentUploadRequest represents a request to upload a new document
type DocumentUploadRequest struct {
	OrganizationID int32                  `json:"organization_id"`
	Title          string                 `json:"title"`
	FileName       string                 `json:"file_name"`
	ContentType    string                 `json:"content_type"`
	FileSize       int64                  `json:"file_size"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// DocumentFilter represents filter options for listing documents
type DocumentFilter struct {
	Status *DocumentStatus `json:"status,omitempty"`
}

// DocumentStats represents document statistics
type DocumentStats struct {
	TotalCount     int64 `json:"total_count"`
	PendingCount   int64 `json:"pending_count"`
	ProcessedCount int64 `json:"processed_count"`
	FailedCount    int64 `json:"failed_count"`
}
