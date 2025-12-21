package domain

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/moasq/go-b2b-starter/internal/files"
)

type FileService interface {
	UploadFile(ctx context.Context, req *FileUploadRequest, content io.Reader) (*FileAsset, error)
	DownloadFile(ctx context.Context, id int32) (io.ReadCloser, *FileAsset, error)
	GetFile(ctx context.Context, id int32) (*FileAsset, error)
	DeleteFile(ctx context.Context, id int32) error
	ListFiles(ctx context.Context, filter *FileSearchFilter, limit, offset int) ([]*FileAsset, error)
	GetFileURL(ctx context.Context, id int32, expiryHours int) (string, error)
}

type fileService struct {
	repo FileRepository
}

func NewFileService(repo FileRepository) FileService {
	return &fileService{
		repo: repo,
	}
}

func (s *fileService) UploadFile(ctx context.Context, req *FileUploadRequest, content io.Reader) (*FileAsset, error) {
	// SECURITY: Sanitize filename to prevent path traversal and dangerous characters
	sanitizedFilename := SanitizeFilename(req.Filename)

	// SECURITY: Validate file extension is allowed
	if !files.IsAllowedFileType(sanitizedFilename) {
		return nil, fmt.Errorf("file type not allowed: %s", sanitizedFilename)
	}

	// Get file category
	category := files.GetFileCategory(sanitizedFilename)

	// SECURITY: Check file size limits
	maxSize := files.GetMaxFileSize(category)
	if req.Size > maxSize {
		return nil, fmt.Errorf("file size %d exceeds limit %d for category %s", req.Size, maxSize, category)
	}

	// SOLUTION: Read entire file into buffer (makes it seekable for R2 retries)
	// AWS SDK v2 requires io.ReadSeeker for retryable uploads
	// bytes.Reader implements io.ReadSeeker, allowing the SDK to rewind the stream
	fileData, err := io.ReadAll(content)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	// Verify actual size matches declared size
	if int64(len(fileData)) != req.Size {
		return nil, fmt.Errorf("file size mismatch: declared %d bytes, actual %d bytes",
			req.Size, len(fileData))
	}

	// SECURITY: Validate file content matches declared extension using magic bytes
	validationReader := bytes.NewReader(fileData)
	if err := ValidateFileContent(validationReader, sanitizedFilename); err != nil {
		return nil, fmt.Errorf("file validation failed: %w", err)
	}

	// Create file asset
	fileAsset := &FileAsset{
		Filename:         sanitizedFilename,
		OriginalFilename: req.Filename, // Keep original for reference
		Size:             req.Size,
		ContentType:      req.ContentType,
		Category:         category,
		Context:          req.Context,
		Metadata:         req.Metadata,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// SOLUTION: Create seekable reader for R2 upload (supports AWS SDK retries)
	seekableContent := bytes.NewReader(fileData)

	// Upload to storage
	if err := s.repo.Upload(ctx, fileAsset, seekableContent); err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return fileAsset, nil
}

func (s *fileService) DownloadFile(ctx context.Context, id int32) (io.ReadCloser, *FileAsset, error) {
	content, fileAsset, err := s.repo.Download(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to download file: %w", err)
	}

	return content, fileAsset, nil
}

func (s *fileService) GetFile(ctx context.Context, id int32) (*FileAsset, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *fileService) DeleteFile(ctx context.Context, id int32) error {
	exists, err := s.repo.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check file existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("file not found")
	}

	return s.repo.Delete(ctx, id)
}

func (s *fileService) ListFiles(ctx context.Context, filter *FileSearchFilter, limit, offset int) ([]*FileAsset, error) {
	return s.repo.List(ctx, filter, limit, offset)
}

func (s *fileService) GetFileURL(ctx context.Context, id int32, expiryHours int) (string, error) {
	fmt.Printf("[FILE-SERVICE] ==============================================\n")
	fmt.Printf("[FILE-SERVICE] GetFileURL requested for file_id=%d, expiry=%dh\n", id, expiryHours)

	fmt.Printf("[FILE-SERVICE] Checking file existence...\n")
	exists, err := s.repo.Exists(ctx, id)
	if err != nil {
		fmt.Printf("[FILE-SERVICE] Exists check failed: %v\n", err)
		fmt.Printf("[FILE-SERVICE] Error type: %T\n", err)
		fmt.Printf("[FILE-SERVICE] ===========================================\n")
		return "", fmt.Errorf("failed to check file existence: %w", err)
	}

	fmt.Printf("[FILE-SERVICE] File exists check result: %v\n", exists)
	if !exists {
		fmt.Printf("[FILE-SERVICE] File not found - returning error\n")
		fmt.Printf("[FILE-SERVICE] This could mean:\n")
		fmt.Printf("  - File ID %d doesn't exist in database\n", id)
		fmt.Printf("  - File exists in database but not in R2 storage\n")
		fmt.Printf("  - Storage path mismatch between database and R2\n")
		fmt.Printf("[FILE-SERVICE] ===========================================\n")
		return "", fmt.Errorf("file not found")
	}

	fmt.Printf("[FILE-SERVICE] File exists, generating \n presigned URL...\n")
	url, err := s.repo.GetURL(ctx, id, expiryHours)
	if err != nil {
		fmt.Printf("[FILE-SERVICE] URL generation failed: %v\n", err)
		fmt.Printf("[FILE-SERVICE] Error type: %T\n", err)
		fmt.Printf("[FILE-SERVICE] ===========================================\n")
		return "", err
	}

	fmt.Printf("[FILE-SERVICE] URL generation successful:\n")
	fmt.Printf("  - URL length: %d characters\n", len(url))
	fmt.Printf("  - URL: %s\n", url)
	fmt.Printf("[FILE-SERVICE] ===========================================\n")
	return url, nil
}

// generateFilePath creates a logical path for organizing files
func generateFilePath(category files.FileCategory, context files.FileContext, filename string) string {
	timestamp := time.Now().Format("2006/01/02")
	return fmt.Sprintf("%s/%s/%s/%s", category, context, timestamp, filename)
}
