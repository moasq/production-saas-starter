-- Documents queries

-- name: CreateDocument :one
INSERT INTO documents.documents (
    organization_id,
    file_asset_id,
    title,
    file_name,
    content_type,
    file_size,
    extracted_text,
    status,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetDocumentByID :one
SELECT * FROM documents.documents
WHERE id = $1 AND organization_id = $2;

-- name: GetDocumentByFileAssetID :one
SELECT * FROM documents.documents
WHERE file_asset_id = $1 AND organization_id = $2;

-- name: ListDocumentsByOrganization :many
SELECT * FROM documents.documents
WHERE organization_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListDocumentsByStatus :many
SELECT * FROM documents.documents
WHERE organization_id = $1 AND status = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: UpdateDocumentStatus :one
UPDATE documents.documents
SET status = $3, updated_at = NOW()
WHERE id = $1 AND organization_id = $2
RETURNING *;

-- name: UpdateDocumentExtractedText :one
UPDATE documents.documents
SET extracted_text = $3, status = 'processed', updated_at = NOW()
WHERE id = $1 AND organization_id = $2
RETURNING *;

-- name: UpdateDocument :one
UPDATE documents.documents
SET
    title = COALESCE($3, title),
    metadata = COALESCE($4, metadata),
    updated_at = NOW()
WHERE id = $1 AND organization_id = $2
RETURNING *;

-- name: DeleteDocument :exec
DELETE FROM documents.documents
WHERE id = $1 AND organization_id = $2;

-- name: CountDocumentsByOrganization :one
SELECT COUNT(*) FROM documents.documents
WHERE organization_id = $1;

-- name: CountDocumentsByStatus :one
SELECT COUNT(*) FROM documents.documents
WHERE organization_id = $1 AND status = $2;
