-- name: CreateFileAsset :one
INSERT INTO file_manager.file_assets (
    file_name,
    original_file_name,
    storage_path,
    bucket_name,
    file_size,
    mime_type,
    file_category_id,
    file_context_id,
    is_public,
    entity_type,
    entity_id,
    purpose,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING *;

-- name: GetFileAssetByID :one
SELECT * FROM file_manager.file_assets
WHERE id = $1;

-- name: DeleteFileAsset :exec
DELETE FROM file_manager.file_assets
WHERE id = $1;

-- name: GetFileAssetsByEntity :many
SELECT * FROM file_manager.file_assets
WHERE entity_type = $1 AND entity_id = $2;

-- name: GetFileAssetsByEntityAndPurpose :many
SELECT * FROM file_manager.file_assets
WHERE entity_type = $1 AND entity_id = $2 AND purpose = $3
ORDER BY created_at DESC;

-- name: GetFileAssetsByCategory :many
SELECT fa.*, fc.name as category_name
FROM file_manager.file_assets fa
JOIN file_manager.file_categories fc ON fa.file_category_id = fc.id  
WHERE fc.name = $1
ORDER BY fa.created_at DESC;

-- name: GetFileAssetsByContext :many
SELECT fa.*, fctx.name as context_name
FROM file_manager.file_assets fa
JOIN file_manager.file_contexts fctx ON fa.file_context_id = fctx.id
WHERE fctx.name = $1
ORDER BY fa.created_at DESC;

-- name: UpdateFileAsset :exec
UPDATE file_manager.file_assets
SET 
    file_name = $2,
    storage_path = $3,
    purpose = $4,
    metadata = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: GetFileAssetByStoragePath :one
SELECT * FROM file_manager.file_assets
WHERE storage_path = $1;

-- name: ListFileAssets :many
SELECT fa.*, fc.name as category_name, fctx.name as context_name
FROM file_manager.file_assets fa
JOIN file_manager.file_categories fc ON fa.file_category_id = fc.id
JOIN file_manager.file_contexts fctx ON fa.file_context_id = fctx.id
ORDER BY fa.created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetFileCategories :many
SELECT * FROM file_manager.file_categories ORDER BY name;

-- name: GetFileContexts :many
SELECT * FROM file_manager.file_contexts ORDER BY name;