package adapterimpl

import (
	"context"

	sqlc "github.com/moasq/go-b2b-starter/pkg/db/postgres/sqlc/gen"
	"github.com/moasq/go-b2b-starter/pkg/db/adapters"
)

// fileAssetStore is a thin wrapper around SQLC store that implements adapters.FileAssetStore
// It only exposes file asset-related methods and delegates directly to the underlying store
type fileAssetStore struct {
	store sqlc.Store
}

func NewFileAssetStore(store sqlc.Store) adapters.FileAssetStore {
	return &fileAssetStore{
		store: store,
	}
}

// Basic file asset operations - direct delegation to SQLC store
func (f *fileAssetStore) CreateFileAsset(ctx context.Context, arg sqlc.CreateFileAssetParams) (sqlc.FileManagerFileAsset, error) {
	return f.store.CreateFileAsset(ctx, arg)
}

func (f *fileAssetStore) GetFileAssetByID(ctx context.Context, id int32) (sqlc.FileManagerFileAsset, error) {
	return f.store.GetFileAssetByID(ctx, id)
}

func (f *fileAssetStore) DeleteFileAsset(ctx context.Context, id int32) error {
	return f.store.DeleteFileAsset(ctx, id)
}

func (f *fileAssetStore) GetFileAssetsByEntity(ctx context.Context, arg sqlc.GetFileAssetsByEntityParams) ([]sqlc.FileManagerFileAsset, error) {
	return f.store.GetFileAssetsByEntity(ctx, arg)
}

func (f *fileAssetStore) GetFileAssetsByEntityAndPurpose(ctx context.Context, arg sqlc.GetFileAssetsByEntityAndPurposeParams) ([]sqlc.FileManagerFileAsset, error) {
	return f.store.GetFileAssetsByEntityAndPurpose(ctx, arg)
}

// Category and context-based operations - direct delegation
func (f *fileAssetStore) GetFileAssetsByCategory(ctx context.Context, categoryName string) ([]sqlc.GetFileAssetsByCategoryRow, error) {
	return f.store.GetFileAssetsByCategory(ctx, categoryName)
}

func (f *fileAssetStore) GetFileAssetsByContext(ctx context.Context, contextName string) ([]sqlc.GetFileAssetsByContextRow, error) {
	return f.store.GetFileAssetsByContext(ctx, contextName)
}

// Update operations - direct delegation
func (f *fileAssetStore) UpdateFileAsset(ctx context.Context, arg sqlc.UpdateFileAssetParams) error {
	return f.store.UpdateFileAsset(ctx, arg)
}

// Search and lookup operations - direct delegation
func (f *fileAssetStore) GetFileAssetByStoragePath(ctx context.Context, storagePath string) (sqlc.FileManagerFileAsset, error) {
	return f.store.GetFileAssetByStoragePath(ctx, storagePath)
}

func (f *fileAssetStore) ListFileAssets(ctx context.Context, arg sqlc.ListFileAssetsParams) ([]sqlc.ListFileAssetsRow, error) {
	return f.store.ListFileAssets(ctx, arg)
}

// Lookup tables operations - direct delegation
func (f *fileAssetStore) GetFileCategories(ctx context.Context) ([]sqlc.FileManagerFileCategory, error) {
	return f.store.GetFileCategories(ctx)
}

func (f *fileAssetStore) GetFileContexts(ctx context.Context) ([]sqlc.FileManagerFileContext, error) {
	return f.store.GetFileContexts(ctx)
}
