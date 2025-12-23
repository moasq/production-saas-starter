package events

import (
	"time"

	"github.com/google/uuid"
	"github.com/moasq/go-b2b-starter/internal/platform/eventbus"
)

const (
	DocumentUploadedEventType  = "document.uploaded"
	DocumentProcessedEventType = "document.processed"
	DocumentFailedEventType    = "document.failed"
)

// DocumentUploaded is published when a document has been uploaded and text extracted
type DocumentUploaded struct {
	eventbus.BaseEvent
	DocumentID     int32  `json:"document_id"`
	OrganizationID int32  `json:"organization_id"`
	FileAssetID    int32  `json:"file_asset_id"`
	Title          string `json:"title"`
	ExtractedText  string `json:"extracted_text"`
}

func NewDocumentUploaded(documentID, organizationID, fileAssetID int32, title, extractedText string) *DocumentUploaded {
	return &DocumentUploaded{
		BaseEvent: eventbus.BaseEvent{
			ID:        uuid.New().String(),
			Name:      DocumentUploadedEventType,
			CreatedAt: time.Now(),
			Meta:      make(map[string]interface{}),
		},
		DocumentID:     documentID,
		OrganizationID: organizationID,
		FileAssetID:    fileAssetID,
		Title:          title,
		ExtractedText:  extractedText,
	}
}

// DocumentProcessed is published when a document embedding has been created
type DocumentProcessed struct {
	eventbus.BaseEvent
	DocumentID     int32 `json:"document_id"`
	OrganizationID int32 `json:"organization_id"`
	EmbeddingID    int32 `json:"embedding_id"`
}

func NewDocumentProcessed(documentID, organizationID, embeddingID int32) *DocumentProcessed {
	return &DocumentProcessed{
		BaseEvent: eventbus.BaseEvent{
			ID:        uuid.New().String(),
			Name:      DocumentProcessedEventType,
			CreatedAt: time.Now(),
			Meta:      make(map[string]interface{}),
		},
		DocumentID:     documentID,
		OrganizationID: organizationID,
		EmbeddingID:    embeddingID,
	}
}

// DocumentFailed is published when document processing fails
type DocumentFailed struct {
	eventbus.BaseEvent
	DocumentID     int32  `json:"document_id"`
	OrganizationID int32  `json:"organization_id"`
	Error          string `json:"error"`
}

func NewDocumentFailed(documentID, organizationID int32, err string) *DocumentFailed {
	return &DocumentFailed{
		BaseEvent: eventbus.BaseEvent{
			ID:        uuid.New().String(),
			Name:      DocumentFailedEventType,
			CreatedAt: time.Now(),
			Meta:      make(map[string]interface{}),
		},
		DocumentID:     documentID,
		OrganizationID: organizationID,
		Error:          err,
	}
}
