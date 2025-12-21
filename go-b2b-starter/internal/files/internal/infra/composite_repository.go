package infra

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/moasq/go-b2b-starter/internal/files/config"
	"github.com/moasq/go-b2b-starter/internal/files/domain"
	file_manager "github.com/moasq/go-b2b-starter/internal/files"
)

type compositeRepository struct {
	r2Repo       domain.R2Repository
	metadataRepo domain.FileMetadataRepository
	bucketName   string
}

func NewCompositeRepository(cfg *config.Config, r2Repo domain.R2Repository, metadataRepo domain.FileMetadataRepository) domain.FileRepository {
	return &compositeRepository{
		r2Repo:       r2Repo,
		metadataRepo: metadataRepo,
		bucketName:   cfg.R2.BucketName,
	}
}

func (r *compositeRepository) Upload(ctx context.Context, file *domain.FileAsset, content io.Reader) error {
	fmt.Printf("[UPLOAD-DEBUG] ==============================================\n")
	fmt.Printf("[UPLOAD-DEBUG] Starting file upload process\n")
	fmt.Printf("[UPLOAD-DEBUG] File details:\n")
	fmt.Printf("  - Filename: %s\n", file.Filename)
	fmt.Printf("  - Size: %d bytes\n", file.Size)
	fmt.Printf("  - Content Type: %s\n", file.ContentType)
	fmt.Printf("  - Category: %s\n", file.Category)
	fmt.Printf("  - Context: %s\n", file.Context)
	
	// Set default values
	file.BucketName = r.bucketName
	file.StoragePath = r.generateStoragePath(file.Category, file.Context, file.Filename)
	
	fmt.Printf("[UPLOAD-DEBUG] Initial configuration:\n")
	fmt.Printf("  - Bucket Name: %s\n", file.BucketName)
	fmt.Printf("  - Initial Storage Path: %s\n", file.StoragePath)
	
	// First, save metadata to get database ID
	fmt.Printf("[UPLOAD-DEBUG] Step 1: Saving metadata to database...\n")
	savedFile, err := r.metadataRepo.Create(ctx, file)
	if err != nil {
		fmt.Printf("[UPLOAD-ERROR] Failed to save file metadata: %v\n", err)
		return fmt.Errorf("failed to save file metadata: %w", err)
	}
	
	fmt.Printf("[UPLOAD-DEBUG] Metadata saved successfully:\n")
	fmt.Printf("  - Assigned File ID: %d\n", savedFile.ID)
	fmt.Printf("  - Database Storage Path: %s\n", savedFile.StoragePath)
	
	// Use database ID as part of the R2 object key
	objectKey := r.generateObjectKey(savedFile.ID, savedFile.Filename)
	fmt.Printf("[UPLOAD-DEBUG] Step 2: Generated R2 object key: %s\n", objectKey)

	// Upload to R2
	fmt.Printf("[UPLOAD-DEBUG] Step 3: Uploading to R2...\n")
	fmt.Printf("  - Bucket: %s\n", r.bucketName)
	fmt.Printf("  - Object Key: %s\n", objectKey)
	fmt.Printf("  - File Size: %d bytes\n", file.Size)
	fmt.Printf("  - Content Type: %s\n", file.ContentType)
	
	err = r.r2Repo.UploadObject(ctx, objectKey, content, file.Size, file.ContentType)
	if err != nil {
		fmt.Printf("[UPLOAD-ERROR] R2 upload failed: %v\n", err)
		fmt.Printf("[UPLOAD-ERROR] Rolling back database entry...\n")
		// Rollback: delete metadata if R2 upload fails
		r.metadataRepo.Delete(ctx, savedFile.ID)
		return fmt.Errorf("failed to upload file to R2: %w", err)
	}

	fmt.Printf("[UPLOAD-DEBUG] R2 upload successful\n")
	
	// Update storage path with the actual object key
	fmt.Printf("[UPLOAD-DEBUG] Step 4: Updating database storage path...\n")
	fmt.Printf("  - Old Path: %s\n", savedFile.StoragePath)
	fmt.Printf("  - New Path: %s\n", objectKey)
	
	savedFile.StoragePath = objectKey
	err = r.metadataRepo.Update(ctx, savedFile)
	if err != nil {
		fmt.Printf("[UPLOAD-ERROR] Database storage path update failed: %v\n", err)
		fmt.Printf("[UPLOAD-ERROR] Error type: %T\n", err)
		fmt.Printf("[UPLOAD-ERROR] Rolling back R2 and database...\n")
		// Rollback: delete from R2 and metadata
		r.r2Repo.DeleteObject(ctx, objectKey)
		r.metadataRepo.Delete(ctx, savedFile.ID)
		return fmt.Errorf("failed to update storage path: %w", err)
	}
	
	fmt.Printf("[UPLOAD-SUCCESS] Database storage path updated successfully\n")
	
	// Update the original file with saved data
	*file = *savedFile
	
	fmt.Printf("[UPLOAD-SUCCESS] File upload completed successfully:\n")
	fmt.Printf("  - File ID: %d\n", savedFile.ID)
	fmt.Printf("  - Final Storage Path: %s\n", savedFile.StoragePath)
	fmt.Printf("  - Bucket: %s\n", savedFile.BucketName)
	fmt.Printf("[UPLOAD-SUCCESS] ============================================\n")
	
	return nil
}

func (r *compositeRepository) Download(ctx context.Context, id int32) (io.ReadCloser, *domain.FileAsset, error) {
	// Get file metadata
	file, err := r.metadataRepo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	// Download from R2
	content, err := r.r2Repo.DownloadObject(ctx, file.StoragePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to download file from R2: %w", err)
	}

	return content, file, nil
}

func (r *compositeRepository) GetByID(ctx context.Context, id int32) (*domain.FileAsset, error) {
	return r.metadataRepo.GetByID(ctx, id)
}

func (r *compositeRepository) Delete(ctx context.Context, id int32) error {
	// Get file metadata first
	file, err := r.metadataRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get file metadata: %w", err)
	}

	// Delete from R2
	err = r.r2Repo.DeleteObject(ctx, file.StoragePath)
	if err != nil {
		return fmt.Errorf("failed to delete file from R2: %w", err)
	}

	// Delete metadata
	err = r.metadataRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete file metadata: %w", err)
	}

	return nil
}

func (r *compositeRepository) List(ctx context.Context, filter *domain.FileSearchFilter, limit, offset int) ([]*domain.FileAsset, error) {
	return r.metadataRepo.List(ctx, filter, limit, offset)
}

func (r *compositeRepository) GetURL(ctx context.Context, id int32, expiryHours int) (string, error) {
	fmt.Printf("[COMPOSITE-REPO] ==============================================\n")
	fmt.Printf("[COMPOSITE-REPO] GetURL requested for file_id=%d, expiry=%dh\n", id, expiryHours)
	
	// Get file metadata
	fmt.Printf("[COMPOSITE-REPO] Fetching file metadata from database...\n")
	file, err := r.metadataRepo.GetByID(ctx, id)
	if err != nil {
		fmt.Printf("[COMPOSITE-REPO] Failed to get file metadata: %v\n", err)
		fmt.Printf("[COMPOSITE-REPO] Error type: %T\n", err)
		fmt.Printf("[COMPOSITE-REPO] ===========================================\n")
		return "", fmt.Errorf("failed to get file metadata: %w", err)
	}
	
	fmt.Printf("[COMPOSITE-REPO] File metadata retrieved:\n")
	fmt.Printf("  - Storage Path: %s\n", file.StoragePath)
	fmt.Printf("  - Bucket Name: %s\n", file.BucketName)
	fmt.Printf("  - File Name: %s\n", file.Filename)

	// Get presigned URL from R2
	fmt.Printf("[COMPOSITE-REPO] Generating R2 presigned URL...\n")
	fmt.Printf("[COMPOSITE-REPO] R2 parameters:\n")
	fmt.Printf("  - Bucket: %s\n", r.bucketName)
	fmt.Printf("  - Object Key: %s\n", file.StoragePath)
	fmt.Printf("  - Expiry: %d hours\n", expiryHours)

	url, err := r.r2Repo.GetPresignedURL(ctx, file.StoragePath, expiryHours)
	if err != nil {
		fmt.Printf("[COMPOSITE-REPO] Failed to get presigned URL: %v\n", err)
		fmt.Printf("[COMPOSITE-REPO] Error type: %T\n", err)
		fmt.Printf("[COMPOSITE-REPO] This could indicate:\n")
		fmt.Printf("  - R2 connection issues\n")
		fmt.Printf("  - Invalid bucket name: %s\n", r.bucketName)
		fmt.Printf("  - Invalid object key: %s\n", file.StoragePath)
		fmt.Printf("  - R2 authentication problems\n")
		fmt.Printf("  - R2 service issues\n")
		fmt.Printf("[COMPOSITE-REPO] ===========================================\n")
		return "", fmt.Errorf("failed to get presigned URL: %w", err)
	}
	
	fmt.Printf("[COMPOSITE-REPO] Presigned URL generated successfully:\n")
	fmt.Printf("  - URL length: %d characters\n", len(url))
	fmt.Printf("  - URL: %s\n", url)
	fmt.Printf("[COMPOSITE-REPO] ===========================================\n")
	
	return url, nil
}

func (r *compositeRepository) Exists(ctx context.Context, id int32) (bool, error) {
	fmt.Printf("[COMPOSITE-REPO] ==============================================\n")
	fmt.Printf("[COMPOSITE-REPO] Checking existence for file_id=%d\n", id)
	
	// Check if metadata exists
	fmt.Printf("[COMPOSITE-REPO] Step 1: Checking file metadata in database...\n")
	file, err := r.metadataRepo.GetByID(ctx, id)
	if err != nil {
		fmt.Printf("[COMPOSITE-REPO] Metadata lookup failed: %v\n", err)
		fmt.Printf("[COMPOSITE-REPO] Error type: %T\n", err)
		fmt.Printf("[COMPOSITE-REPO] This means file_id=%d does not exist in database\n", id)
		fmt.Printf("[COMPOSITE-REPO] ===========================================\n")
		return false, fmt.Errorf("failed to check file metadata: %w", err) // FIX THE BUG
	}
	
	fmt.Printf("[COMPOSITE-REPO] Metadata found successfully:\n")
	fmt.Printf("  - File ID: %d\n", file.ID)
	fmt.Printf("  - Filename: %s\n", file.Filename)
	fmt.Printf("  - Storage Path: %s\n", file.StoragePath)
	fmt.Printf("  - Bucket Name: %s\n", file.BucketName)
	fmt.Printf("  - Content Type: %s\n", file.ContentType)
	fmt.Printf("  - File Size: %d bytes\n", file.Size)

	// Check if object exists in R2
	fmt.Printf("[COMPOSITE-REPO] Step 2: Checking object existence in R2...\n")
	fmt.Printf("[COMPOSITE-REPO] Looking for object: %s in bucket: %s\n", file.StoragePath, r.bucketName)

	exists, err := r.r2Repo.ObjectExists(ctx, file.StoragePath)
	if err != nil {
		fmt.Printf("[COMPOSITE-REPO] R2 existence check failed: %v\n", err)
		fmt.Printf("[COMPOSITE-REPO] Error type: %T\n", err)
		fmt.Printf("[COMPOSITE-REPO] This could indicate:\n")
		fmt.Printf("  - R2 connection problems\n")
		fmt.Printf("  - Incorrect bucket name: %s\n", r.bucketName)
		fmt.Printf("  - Incorrect object path: %s\n", file.StoragePath)
		fmt.Printf("  - R2 authentication issues\n")
		fmt.Printf("[COMPOSITE-REPO] ===========================================\n")
		return false, fmt.Errorf("failed to check R2 object existence: %w", err)
	}

	fmt.Printf("[COMPOSITE-REPO] R2 object existence check result: %v\n", exists)
	if !exists {
		fmt.Printf("[COMPOSITE-REPO] File metadata exists in database but object missing in R2\n")
		fmt.Printf("[COMPOSITE-REPO] Expected object path: %s\n", file.StoragePath)
		fmt.Printf("[COMPOSITE-REPO] Expected bucket: %s\n", r.bucketName)
		fmt.Printf("[COMPOSITE-REPO] This indicates a storage consistency issue\n")
	} else {
		fmt.Printf("[COMPOSITE-REPO] File exists in both database and R2 storage\n")
	}
	
	fmt.Printf("[COMPOSITE-REPO] ===========================================\n")
	return exists, nil
}

func (r *compositeRepository) GetByCategory(ctx context.Context, category files.FileCategory, limit, offset int) ([]*domain.FileAsset, error) {
	return r.metadataRepo.GetByCategory(ctx, string(category), limit, offset)
}

func (r *compositeRepository) GetByContext(ctx context.Context, context files.FileContext, limit, offset int) ([]*domain.FileAsset, error) {
	return r.metadataRepo.GetByContext(ctx, string(context), limit, offset)
}

func (r *compositeRepository) GetByEntity(ctx context.Context, entityType string, entityID int32) ([]*domain.FileAsset, error) {
	return r.metadataRepo.GetByEntity(ctx, entityType, entityID)
}

// Helper methods
func (r *compositeRepository) generateStoragePath(category files.FileCategory, context files.FileContext, filename string) string {
	timestamp := time.Now().Format("2006/01/02")
	return fmt.Sprintf("%s/%s/%s/%s", category, context, timestamp, filename)
}

func (r *compositeRepository) generateObjectKey(id int32, filename string) string {
	// Use database ID as part of the object key for easy lookup
	return fmt.Sprintf("files/%d/%s", id, filename)
}