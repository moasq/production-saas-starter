package infra

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	sqlc "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
	"github.com/moasq/go-b2b-starter/internal/db/adapters"
	file_manager "github.com/moasq/go-b2b-starter/internal/files"
	"github.com/moasq/go-b2b-starter/internal/files/domain"
)

type dbRepository struct {
	store adapters.FileAssetStore
}

func NewDBRepository(store adapters.FileAssetStore) domain.FileMetadataRepository {
	return &dbRepository{
		store: store,
	}
}

func (r *dbRepository) Create(ctx context.Context, file *domain.FileAsset) (*domain.FileAsset, error) {
	// Convert metadata map to JSON bytes
	metadataBytes, err := json.Marshal(file.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Get category and context IDs
	categoryID, err := r.getCategoryID(ctx, file.Category)
	if err != nil {
		return nil, fmt.Errorf("failed to get category ID: %w", err)
	}

	contextID, err := r.getContextID(ctx, file.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to get context ID: %w", err)
	}

	params := sqlc.CreateFileAssetParams{
		FileName:         file.Filename,
		OriginalFileName: file.OriginalFilename,
		StoragePath:      file.StoragePath,
		BucketName:       file.BucketName,
		FileSize:         file.Size,
		MimeType:         file.ContentType,
		FileCategoryID:   categoryID,
		FileContextID:    contextID,
		IsPublic:         pgtype.Bool{Bool: file.IsPublic, Valid: true},
		EntityType:       pgtype.Text{String: file.EntityType, Valid: file.EntityType != ""},
		EntityID:         pgtype.Int4{Int32: file.EntityID, Valid: file.EntityID != 0},
		Purpose:          pgtype.Text{String: file.Purpose, Valid: file.Purpose != ""},
		Metadata:         metadataBytes,
	}

	dbFile, err := r.store.CreateFileAsset(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create file asset: %w", err)
	}

	return r.convertFromDBModel(&dbFile), nil
}

func (r *dbRepository) GetByID(ctx context.Context, id int32) (*domain.FileAsset, error) {
	fmt.Printf("[DB-REPO] ==============================================\n")
	fmt.Printf("[DB-REPO] Querying file_asset table for id=%d\n", id)

	dbFile, err := r.store.GetFileAssetByID(ctx, id)
	if err != nil {
		fmt.Printf("[DB-REPO] Database query failed: %v\n", err)
		fmt.Printf("[DB-REPO] Error type: %T\n", err)
		fmt.Printf("[DB-REPO] This could mean:\n")
		fmt.Printf("  - File ID %d does not exist in database\n", id)
		fmt.Printf("  - Database connection problems\n")
		fmt.Printf("  - SQL query execution issues\n")
		fmt.Printf("  - Table or column structure problems\n")
		fmt.Printf("[DB-REPO] ===========================================\n")
		return nil, fmt.Errorf("failed to get file asset: %w", err)
	}

	fmt.Printf("[DB-REPO] File found in database:\n")
	fmt.Printf("  - ID: %d\n", dbFile.ID)
	fmt.Printf("  - Filename: %s\n", dbFile.FileName)
	fmt.Printf("  - Original Filename: %s\n", dbFile.OriginalFileName)
	fmt.Printf("  - Storage Path: %s\n", dbFile.StoragePath)
	fmt.Printf("  - Bucket Name: %s\n", dbFile.BucketName)
	fmt.Printf("  - File Size: %d bytes\n", dbFile.FileSize)
	fmt.Printf("  - MIME Type: %s\n", dbFile.MimeType)
	fmt.Printf("  - Created At: %v\n", dbFile.CreatedAt.Time)
	fmt.Printf("  - Updated At: %v\n", dbFile.UpdatedAt.Time)

	file := r.convertFromDBModel(&dbFile)
	fmt.Printf("[DB-REPO] File conversion successful\n")
	fmt.Printf("[DB-REPO] ===========================================\n")

	return file, nil
}

func (r *dbRepository) Update(ctx context.Context, file *domain.FileAsset) error {
	fmt.Printf("[DB-UPDATE] ==============================================\n")
	fmt.Printf("[DB-UPDATE] Updating file asset in database\n")
	fmt.Printf("[DB-UPDATE] File details:\n")
	fmt.Printf("  - File ID: %d\n", file.ID)
	fmt.Printf("  - Filename: %s\n", file.Filename)
	fmt.Printf("  - Storage Path: %s\n", file.StoragePath)
	fmt.Printf("  - Purpose: %s\n", file.Purpose)

	metadataBytes, err := json.Marshal(file.Metadata)
	if err != nil {
		fmt.Printf("[DB-UPDATE-ERROR] Failed to marshal metadata: %v\n", err)
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	params := sqlc.UpdateFileAssetParams{
		ID:          file.ID,
		FileName:    file.Filename,
		StoragePath: file.StoragePath, // FIX: Add missing StoragePath field
		Purpose:     pgtype.Text{String: file.Purpose, Valid: file.Purpose != ""},
		Metadata:    metadataBytes,
	}

	fmt.Printf("[DB-UPDATE] Executing database update...\n")
	fmt.Printf("  - Updating Storage Path to: %s\n", file.StoragePath)

	err = r.store.UpdateFileAsset(ctx, params)
	if err != nil {
		fmt.Printf("[DB-UPDATE-ERROR] Database update failed: %v\n", err)
		fmt.Printf("[DB-UPDATE-ERROR] Error type: %T\n", err)
		return err
	}

	fmt.Printf("[DB-UPDATE-SUCCESS] Database update completed successfully\n")
	fmt.Printf("[DB-UPDATE-SUCCESS] ==========================================\n")

	return nil
}

func (r *dbRepository) Delete(ctx context.Context, id int32) error {
	return r.store.DeleteFileAsset(ctx, id)
}

func (r *dbRepository) List(ctx context.Context, filter *domain.FileSearchFilter, limit, offset int) ([]*domain.FileAsset, error) {
	params := sqlc.ListFileAssetsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	rows, err := r.store.ListFileAssets(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list file assets: %w", err)
	}

	files := make([]*domain.FileAsset, len(rows))
	for i, row := range rows {
		files[i] = r.convertFromListRow(&row)
	}

	return files, nil
}

func (r *dbRepository) GetByStoragePath(ctx context.Context, storagePath string) (*domain.FileAsset, error) {
	dbFile, err := r.store.GetFileAssetByStoragePath(ctx, storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file asset by storage path: %w", err)
	}

	return r.convertFromDBModel(&dbFile), nil
}

func (r *dbRepository) GetByCategory(ctx context.Context, category string, limit, offset int) ([]*domain.FileAsset, error) {
	rows, err := r.store.GetFileAssetsByCategory(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("failed to get file assets by category: %w", err)
	}

	files := make([]*domain.FileAsset, len(rows))
	for i, row := range rows {
		files[i] = r.convertFromCategoryRow(&row)
	}

	return files, nil
}

func (r *dbRepository) GetByContext(ctx context.Context, context string, limit, offset int) ([]*domain.FileAsset, error) {
	rows, err := r.store.GetFileAssetsByContext(ctx, context)
	if err != nil {
		return nil, fmt.Errorf("failed to get file assets by context: %w", err)
	}

	files := make([]*domain.FileAsset, len(rows))
	for i, row := range rows {
		files[i] = r.convertFromContextRow(&row)
	}

	return files, nil
}

func (r *dbRepository) GetByEntity(ctx context.Context, entityType string, entityID int32) ([]*domain.FileAsset, error) {
	params := sqlc.GetFileAssetsByEntityParams{
		EntityType: pgtype.Text{String: entityType, Valid: true},
		EntityID:   pgtype.Int4{Int32: entityID, Valid: true},
	}

	dbFiles, err := r.store.GetFileAssetsByEntity(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get file assets by entity: %w", err)
	}

	files := make([]*domain.FileAsset, len(dbFiles))
	for i, dbFile := range dbFiles {
		files[i] = r.convertFromDBModel(&dbFile)
	}

	return files, nil
}

// Helper methods for conversion and lookup
func (r *dbRepository) getCategoryID(ctx context.Context, category files.FileCategory) (int16, error) {
	categories, err := r.store.GetFileCategories(ctx)
	if err != nil {
		return 0, err
	}

	for _, cat := range categories {
		if cat.Name == string(category) {
			return cat.ID, nil
		}
	}

	return 0, fmt.Errorf("category not found: %s", category)
}

func (r *dbRepository) getContextID(ctx context.Context, context files.FileContext) (int16, error) {
	contexts, err := r.store.GetFileContexts(ctx)
	if err != nil {
		return 0, err
	}

	for _, ctx := range contexts {
		if ctx.Name == string(context) {
			return ctx.ID, nil
		}
	}

	return 0, fmt.Errorf("context not found: %s", context)
}

func (r *dbRepository) convertFromDBModel(dbFile *sqlc.FileManagerFileAsset) *domain.FileAsset {
	var metadata map[string]interface{}
	if len(dbFile.Metadata) > 0 {
		json.Unmarshal(dbFile.Metadata, &metadata)
	}

	var entityType string
	if dbFile.EntityType.Valid {
		entityType = dbFile.EntityType.String
	}

	var entityID int32
	if dbFile.EntityID.Valid {
		entityID = dbFile.EntityID.Int32
	}

	var purpose string
	if dbFile.Purpose.Valid {
		purpose = dbFile.Purpose.String
	}

	var isPublic bool
	if dbFile.IsPublic.Valid {
		isPublic = dbFile.IsPublic.Bool
	}

	return &domain.FileAsset{
		ID:               dbFile.ID,
		UUID:             uuid.New(), // Generate UUID for external reference
		Filename:         dbFile.FileName,
		OriginalFilename: dbFile.OriginalFileName,
		Size:             dbFile.FileSize,
		ContentType:      dbFile.MimeType,
		StoragePath:      dbFile.StoragePath,
		BucketName:       dbFile.BucketName,
		IsPublic:         isPublic,
		EntityType:       entityType,
		EntityID:         entityID,
		Purpose:          purpose,
		Metadata:         metadata,
		CreatedAt:        dbFile.CreatedAt.Time,
		UpdatedAt:        dbFile.UpdatedAt.Time,
	}
}

func (r *dbRepository) convertFromListRow(row *sqlc.ListFileAssetsRow) *domain.FileAsset {
	var metadata map[string]interface{}
	if len(row.Metadata) > 0 {
		json.Unmarshal(row.Metadata, &metadata)
	}

	var entityType string
	if row.EntityType.Valid {
		entityType = row.EntityType.String
	}

	var entityID int32
	if row.EntityID.Valid {
		entityID = row.EntityID.Int32
	}

	var purpose string
	if row.Purpose.Valid {
		purpose = row.Purpose.String
	}

	var isPublic bool
	if row.IsPublic.Valid {
		isPublic = row.IsPublic.Bool
	}

	return &domain.FileAsset{
		ID:               row.ID,
		UUID:             uuid.New(),
		Filename:         row.FileName,
		OriginalFilename: row.OriginalFileName,
		Size:             row.FileSize,
		ContentType:      row.MimeType,
		Category:         files.FileCategory(row.CategoryName),
		Context:          files.FileContext(row.ContextName),
		StoragePath:      row.StoragePath,
		BucketName:       row.BucketName,
		IsPublic:         isPublic,
		EntityType:       entityType,
		EntityID:         entityID,
		Purpose:          purpose,
		Metadata:         metadata,
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

func (r *dbRepository) convertFromCategoryRow(row *sqlc.GetFileAssetsByCategoryRow) *domain.FileAsset {
	var metadata map[string]interface{}
	if len(row.Metadata) > 0 {
		json.Unmarshal(row.Metadata, &metadata)
	}

	var entityType string
	if row.EntityType.Valid {
		entityType = row.EntityType.String
	}

	var entityID int32
	if row.EntityID.Valid {
		entityID = row.EntityID.Int32
	}

	var purpose string
	if row.Purpose.Valid {
		purpose = row.Purpose.String
	}

	var isPublic bool
	if row.IsPublic.Valid {
		isPublic = row.IsPublic.Bool
	}

	return &domain.FileAsset{
		ID:               row.ID,
		UUID:             uuid.New(),
		Filename:         row.FileName,
		OriginalFilename: row.OriginalFileName,
		Size:             row.FileSize,
		ContentType:      row.MimeType,
		Category:         files.FileCategory(row.CategoryName),
		StoragePath:      row.StoragePath,
		BucketName:       row.BucketName,
		IsPublic:         isPublic,
		EntityType:       entityType,
		EntityID:         entityID,
		Purpose:          purpose,
		Metadata:         metadata,
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}

func (r *dbRepository) convertFromContextRow(row *sqlc.GetFileAssetsByContextRow) *domain.FileAsset {
	var metadata map[string]interface{}
	if len(row.Metadata) > 0 {
		json.Unmarshal(row.Metadata, &metadata)
	}

	var entityType string
	if row.EntityType.Valid {
		entityType = row.EntityType.String
	}

	var entityID int32
	if row.EntityID.Valid {
		entityID = row.EntityID.Int32
	}

	var purpose string
	if row.Purpose.Valid {
		purpose = row.Purpose.String
	}

	var isPublic bool
	if row.IsPublic.Valid {
		isPublic = row.IsPublic.Bool
	}

	return &domain.FileAsset{
		ID:               row.ID,
		UUID:             uuid.New(),
		Filename:         row.FileName,
		OriginalFilename: row.OriginalFileName,
		Size:             row.FileSize,
		ContentType:      row.MimeType,
		Context:          files.FileContext(row.ContextName),
		StoragePath:      row.StoragePath,
		BucketName:       row.BucketName,
		IsPublic:         isPublic,
		EntityType:       entityType,
		EntityID:         entityID,
		Purpose:          purpose,
		Metadata:         metadata,
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}
}
