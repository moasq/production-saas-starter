package adapters

import (
	"context"
	
	db "github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc/gen"
)

// FileAssetStore defines the interface for file asset database operations
// It exposes only file asset-related methods and returns SQLC types directly
type FileAssetStore interface {
	// Basic file asset operations - using SQLC method signatures
	CreateFileAsset(ctx context.Context, arg db.CreateFileAssetParams) (db.FileManagerFileAsset, error)
	GetFileAssetByID(ctx context.Context, id int32) (db.FileManagerFileAsset, error)
	DeleteFileAsset(ctx context.Context, id int32) error
	GetFileAssetsByEntity(ctx context.Context, arg db.GetFileAssetsByEntityParams) ([]db.FileManagerFileAsset, error)
	GetFileAssetsByEntityAndPurpose(ctx context.Context, arg db.GetFileAssetsByEntityAndPurposeParams) ([]db.FileManagerFileAsset, error)
	
	// Category and context-based operations
	GetFileAssetsByCategory(ctx context.Context, categoryName string) ([]db.GetFileAssetsByCategoryRow, error)
	GetFileAssetsByContext(ctx context.Context, contextName string) ([]db.GetFileAssetsByContextRow, error)
	
	// Update operations
	UpdateFileAsset(ctx context.Context, arg db.UpdateFileAssetParams) error
	
	// Search and lookup operations
	GetFileAssetByStoragePath(ctx context.Context, storagePath string) (db.FileManagerFileAsset, error)
	ListFileAssets(ctx context.Context, arg db.ListFileAssetsParams) ([]db.ListFileAssetsRow, error)
	
	// Lookup tables operations
	GetFileCategories(ctx context.Context) ([]db.FileManagerFileCategory, error)
	GetFileContexts(ctx context.Context) ([]db.FileManagerFileContext, error)
}