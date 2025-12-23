package infra

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	file_manager "github.com/moasq/go-b2b-starter/internal/modules/files"
	"github.com/moasq/go-b2b-starter/internal/modules/files/domain"
	sqlc "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
)

// fileMetadataRepository implements domain.FileMetadataRepository using SQLC internally.
// SQLC types are never exposed outside this package.
type fileMetadataRepository struct {
	store sqlc.Store
}

// NewFileMetadataRepository creates a new FileMetadataRepository implementation.
func NewFileMetadataRepository(store sqlc.Store) domain.FileMetadataRepository {
	return &fileMetadataRepository{store: store}
}

func (r *fileMetadataRepository) Create(ctx context.Context, file *domain.FileAsset) (*domain.FileAsset, error) {
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

func (r *fileMetadataRepository) GetByID(ctx context.Context, id int32) (*domain.FileAsset, error) {
	dbFile, err := r.store.GetFileAssetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get file asset: %w", err)
	}

	return r.convertFromDBModel(&dbFile), nil
}

func (r *fileMetadataRepository) Update(ctx context.Context, file *domain.FileAsset) error {
	metadataBytes, err := json.Marshal(file.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	params := sqlc.UpdateFileAssetParams{
		ID:          file.ID,
		FileName:    file.Filename,
		StoragePath: file.StoragePath,
		Purpose:     pgtype.Text{String: file.Purpose, Valid: file.Purpose != ""},
		Metadata:    metadataBytes,
	}

	return r.store.UpdateFileAsset(ctx, params)
}

func (r *fileMetadataRepository) Delete(ctx context.Context, id int32) error {
	return r.store.DeleteFileAsset(ctx, id)
}

func (r *fileMetadataRepository) List(ctx context.Context, filter *domain.FileSearchFilter, limit, offset int) ([]*domain.FileAsset, error) {
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

func (r *fileMetadataRepository) GetByStoragePath(ctx context.Context, storagePath string) (*domain.FileAsset, error) {
	dbFile, err := r.store.GetFileAssetByStoragePath(ctx, storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file asset by storage path: %w", err)
	}

	return r.convertFromDBModel(&dbFile), nil
}

func (r *fileMetadataRepository) GetByCategory(ctx context.Context, category string, limit, offset int) ([]*domain.FileAsset, error) {
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

func (r *fileMetadataRepository) GetByContext(ctx context.Context, fileContext string, limit, offset int) ([]*domain.FileAsset, error) {
	rows, err := r.store.GetFileAssetsByContext(ctx, fileContext)
	if err != nil {
		return nil, fmt.Errorf("failed to get file assets by context: %w", err)
	}

	files := make([]*domain.FileAsset, len(rows))
	for i, row := range rows {
		files[i] = r.convertFromContextRow(&row)
	}

	return files, nil
}

func (r *fileMetadataRepository) GetByEntity(ctx context.Context, entityType string, entityID int32) ([]*domain.FileAsset, error) {
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

func (r *fileMetadataRepository) getCategoryID(ctx context.Context, category file_manager.FileCategory) (int16, error) {
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

func (r *fileMetadataRepository) getContextID(ctx context.Context, fileContext file_manager.FileContext) (int16, error) {
	contexts, err := r.store.GetFileContexts(ctx)
	if err != nil {
		return 0, err
	}

	for _, ctx := range contexts {
		if ctx.Name == string(fileContext) {
			return ctx.ID, nil
		}
	}

	return 0, fmt.Errorf("context not found: %s", fileContext)
}

// Conversion functions - translate SQLC types to domain types

func (r *fileMetadataRepository) convertFromDBModel(dbFile *sqlc.FileManagerFileAsset) *domain.FileAsset {
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
		UUID:             uuid.New(),
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

func (r *fileMetadataRepository) convertFromListRow(row *sqlc.ListFileAssetsRow) *domain.FileAsset {
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
		Category:         file_manager.FileCategory(row.CategoryName),
		Context:          file_manager.FileContext(row.ContextName),
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

func (r *fileMetadataRepository) convertFromCategoryRow(row *sqlc.GetFileAssetsByCategoryRow) *domain.FileAsset {
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
		Category:         file_manager.FileCategory(row.CategoryName),
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

func (r *fileMetadataRepository) convertFromContextRow(row *sqlc.GetFileAssetsByContextRow) *domain.FileAsset {
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
		Context:          file_manager.FileContext(row.ContextName),
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
