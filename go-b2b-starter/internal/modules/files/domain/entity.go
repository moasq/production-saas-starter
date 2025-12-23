package domain

import (
	"time"

	"github.com/moasq/go-b2b-starter/internal/modules/files"
	"github.com/google/uuid"
)

type FileAsset struct {
	ID               int32                     `json:"id"`   // Database ID
	UUID             uuid.UUID                 `json:"uuid"` // UUID for external reference
	Filename         string                    `json:"filename"`
	OriginalFilename string                    `json:"original_filename"`
	Size             int64                     `json:"size"`
	ContentType      string                    `json:"content_type"`
	Category         files.FileCategory `json:"category"`
	Context          files.FileContext  `json:"context"`
	StoragePath      string                    `json:"storage_path"` // R2 object path
	BucketName       string                    `json:"bucket_name"`
	IsPublic         bool                      `json:"is_public"`
	EntityType       string                    `json:"entity_type,omitempty"`
	EntityID         int32                     `json:"entity_id,omitempty"`
	Purpose          string                    `json:"purpose,omitempty"`
	Metadata         map[string]interface{}    `json:"metadata,omitempty"`
	URL              string                    `json:"url,omitempty"` // Presigned URL
	CreatedAt        time.Time                 `json:"created_at"`
	UpdatedAt        time.Time                 `json:"updated_at"`
}

type FileUploadRequest struct {
	Filename    string                   `json:"filename"`
	Size        int64                    `json:"size"`
	ContentType string                   `json:"content_type"`
	Context     files.FileContext `json:"context"`
	Metadata    map[string]any           `json:"metadata,omitempty"`
}

type FileSearchFilter struct {
	Category *files.FileCategory `json:"category,omitempty"`
	Context  *files.FileContext  `json:"context,omitempty"`
	MinSize  *int64                     `json:"min_size,omitempty"`
	MaxSize  *int64                     `json:"max_size,omitempty"`
	DateFrom *time.Time                 `json:"date_from,omitempty"`
	DateTo   *time.Time                 `json:"date_to,omitempty"`
}
