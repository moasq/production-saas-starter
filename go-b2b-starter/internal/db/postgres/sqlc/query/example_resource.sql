-- Example Resource Queries
-- Demonstrates Clean Architecture patterns with CRUD operations,
-- file attachments, OCR/LLM processing, and approval workflows

-- CREATE operations

-- name: CreateResource :one
INSERT INTO example_resources (
    resource_number, title, description, status_id, file_id,
    extracted_data, processed_data, confidence,
    organization_id, created_by_account_id, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: CreateMinimalResource :one
-- Creates a minimal placeholder resource
INSERT INTO example_resources (
    resource_number, title, organization_id, created_by_account_id, status_id
) VALUES (
    $1, $2, $3, $4, 1
) RETURNING *;

-- READ operations

-- name: GetResourceByID :one
SELECT * FROM example_resources
WHERE id = $1 AND organization_id = $2 AND is_active = true;

-- name: GetResourceByNumber :one
SELECT * FROM example_resources
WHERE resource_number = $1 AND organization_id = $2 AND is_active = true;

-- name: ListResources :many
-- List resources with filtering and pagination
SELECT
    id, resource_number, title, description, status_id, file_id,
    confidence, organization_id, created_by_account_id,
    approval_status, approval_assigned_to_id,
    is_active, created_at, updated_at
FROM example_resources
WHERE organization_id = $1 AND is_active = true
    AND ($2::smallint IS NULL OR status_id = $2)
    AND ($3::varchar IS NULL OR approval_status = $3)
    AND ($4::text IS NULL OR title ILIKE '%' || $4 || '%' OR description ILIKE '%' || $4 || '%')
ORDER BY created_at DESC
LIMIT $5 OFFSET $6;

-- name: CountResources :one
-- Count resources for pagination
SELECT COUNT(*) FROM example_resources
WHERE organization_id = $1 AND is_active = true
    AND ($2::smallint IS NULL OR status_id = $2)
    AND ($3::varchar IS NULL OR approval_status = $3)
    AND ($4::text IS NULL OR title ILIKE '%' || $4 || '%' OR description ILIKE '%' || $4 || '%');

-- UPDATE operations

-- name: UpdateResource :exec
UPDATE example_resources SET
    title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    status_id = COALESCE(sqlc.narg('status_id'), status_id),
    metadata = COALESCE(sqlc.narg('metadata'), metadata),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id') AND organization_id = sqlc.arg('organization_id') AND is_active = true;

-- name: UpdateResourceProcessingData :exec
-- Update OCR/LLM processing results
UPDATE example_resources SET
    extracted_data = COALESCE(sqlc.narg('extracted_data'), extracted_data),
    processed_data = COALESCE(sqlc.narg('processed_data'), processed_data),
    confidence = COALESCE(sqlc.narg('confidence'), confidence),
    status_id = COALESCE(sqlc.narg('status_id'), status_id),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id') AND organization_id = sqlc.arg('organization_id');

-- name: UpdateResourceStatus :exec
UPDATE example_resources SET
    status_id = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND organization_id = $2 AND is_active = true;

-- name: UpdateResourceApprovalStatus :exec
-- Update approval workflow status
UPDATE example_resources SET
    approval_status = $3,
    approval_action_taker_id = $4,
    approval_notes = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND organization_id = $2 AND is_active = true;

-- name: AssignResourceApproval :exec
-- Assign resource to someone for approval
UPDATE example_resources SET
    approval_assigned_to_id = $3,
    approval_status = 'pending',
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND organization_id = $2 AND is_active = true;

-- name: AttachFileToResource :exec
-- Attach a file to a resource
UPDATE example_resources SET
    file_id = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND organization_id = $2 AND is_active = true;

-- DELETE operations

-- name: DeleteResource :exec
-- Soft delete a resource
UPDATE example_resources SET
    is_active = false,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND organization_id = $2;

-- name: HardDeleteResource :exec
-- Hard delete a resource (use with caution)
DELETE FROM example_resources
WHERE id = $1 AND organization_id = $2;

-- SEARCH operations

-- name: SearchResourcesByText :many
-- Full-text search on title and description
SELECT
    id, resource_number, title, description, status_id,
    confidence, created_at, updated_at,
    ts_rank(to_tsvector('english', coalesce(title, '') || ' ' || coalesce(description, '')), to_tsquery('english', $2)) AS rank
FROM example_resources
WHERE organization_id = $1
    AND is_active = true
    AND to_tsvector('english', coalesce(title, '') || ' ' || coalesce(description, '')) @@ to_tsquery('english', $2)
ORDER BY rank DESC, created_at DESC
LIMIT $3 OFFSET $4;

-- ANALYTICS queries

-- name: GetResourceStats :one
-- Get statistics for dashboard
SELECT
    COUNT(*) as total_resources,
    COUNT(*) FILTER (WHERE status_id = 1) as draft_count,
    COUNT(*) FILTER (WHERE status_id = 2) as processing_count,
    COUNT(*) FILTER (WHERE status_id = 3) as completed_count,
    COUNT(*) FILTER (WHERE approval_status = 'pending') as pending_approval,
    COUNT(*) FILTER (WHERE approval_status = 'approved') as approved_count,
    AVG(confidence) as avg_confidence
FROM example_resources
WHERE organization_id = $1 AND is_active = true;

-- name: GetRecentResources :many
-- Get most recently created resources
SELECT
    id, resource_number, title, status_id, confidence,
    created_by_account_id, created_at
FROM example_resources
WHERE organization_id = $1 AND is_active = true
ORDER BY created_at DESC
LIMIT $2;

-- name: GetResourcesByCreator :many
-- Get resources created by a specific user
SELECT * FROM example_resources
WHERE organization_id = $1
    AND created_by_account_id = $2
    AND is_active = true
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;
