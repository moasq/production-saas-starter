package domain

import (
	"context"
	"io"

	file_manager "github.com/moasq/go-b2b-starter/internal/files"
)

type FileRepository interface {
	// Combined operations (R2 + Database)
	Upload(ctx context.Context, file *FileAsset, content io.Reader) error
	Download(ctx context.Context, id int32) (io.ReadCloser, *FileAsset, error)
	GetByID(ctx context.Context, id int32) (*FileAsset, error)
	Delete(ctx context.Context, id int32) error
	List(ctx context.Context, filter *FileSearchFilter, limit, offset int) ([]*FileAsset, error)
	GetURL(ctx context.Context, id int32, expiryHours int) (string, error)
	Exists(ctx context.Context, id int32) (bool, error)

	// Additional operations
	GetByCategory(ctx context.Context, category files.FileCategory, limit, offset int) ([]*FileAsset, error)
	GetByContext(ctx context.Context, context files.FileContext, limit, offset int) ([]*FileAsset, error)
	GetByEntity(ctx context.Context, entityType string, entityID int32) ([]*FileAsset, error)
}

// R2Repository handles only object storage operations (Cloudflare R2)
type R2Repository interface {
	UploadObject(ctx context.Context, objectKey string, content io.Reader, size int64, contentType string) error
	DownloadObject(ctx context.Context, objectKey string) (io.ReadCloser, error)
	DeleteObject(ctx context.Context, objectKey string) error
	GetPresignedURL(ctx context.Context, objectKey string, expiryHours int) (string, error)
	ObjectExists(ctx context.Context, objectKey string) (bool, error)
}

// FileMetadataRepository handles only database operations
type FileMetadataRepository interface {
	Create(ctx context.Context, file *FileAsset) (*FileAsset, error)
	GetByID(ctx context.Context, id int32) (*FileAsset, error)
	Update(ctx context.Context, file *FileAsset) error
	Delete(ctx context.Context, id int32) error
	List(ctx context.Context, filter *FileSearchFilter, limit, offset int) ([]*FileAsset, error)
	GetByStoragePath(ctx context.Context, storagePath string) (*FileAsset, error)
	GetByCategory(ctx context.Context, category string, limit, offset int) ([]*FileAsset, error)
	GetByContext(ctx context.Context, context string, limit, offset int) ([]*FileAsset, error)
	GetByEntity(ctx context.Context, entityType string, entityID int32) ([]*FileAsset, error)
}
