-- Cognitive Agent queries

-- Document Embeddings

-- name: CreateDocumentEmbedding :one
INSERT INTO cognitive.document_embeddings (
    document_id,
    organization_id,
    embedding,
    content_hash,
    content_preview,
    chunk_index
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetDocumentEmbeddingByID :one
SELECT * FROM cognitive.document_embeddings
WHERE id = $1 AND organization_id = $2;

-- name: GetDocumentEmbeddingsByDocumentID :many
SELECT * FROM cognitive.document_embeddings
WHERE document_id = $1 AND organization_id = $2
ORDER BY chunk_index;

-- name: SearchSimilarDocuments :many
SELECT
    de.id,
    de.document_id,
    de.organization_id,
    de.content_hash,
    de.content_preview,
    de.chunk_index,
    de.created_at,
    de.updated_at,
    (1 - (de.embedding <=> $1::vector))::double precision as similarity_score
FROM cognitive.document_embeddings de
WHERE de.organization_id = $2
ORDER BY de.embedding <=> $1::vector
LIMIT $3;

-- name: DeleteDocumentEmbeddings :exec
DELETE FROM cognitive.document_embeddings
WHERE document_id = $1 AND organization_id = $2;

-- name: CountDocumentEmbeddingsByOrganization :one
SELECT COUNT(*) FROM cognitive.document_embeddings
WHERE organization_id = $1;

-- Chat Sessions

-- name: CreateChatSession :one
INSERT INTO cognitive.chat_sessions (
    organization_id,
    account_id,
    title
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetChatSessionByID :one
SELECT * FROM cognitive.chat_sessions
WHERE id = $1 AND organization_id = $2;

-- name: ListChatSessionsByAccount :many
SELECT * FROM cognitive.chat_sessions
WHERE organization_id = $1 AND account_id = $2
ORDER BY updated_at DESC
LIMIT $3 OFFSET $4;

-- name: UpdateChatSessionTitle :one
UPDATE cognitive.chat_sessions
SET title = $3, updated_at = NOW()
WHERE id = $1 AND organization_id = $2
RETURNING *;

-- name: DeleteChatSession :exec
DELETE FROM cognitive.chat_sessions
WHERE id = $1 AND organization_id = $2;

-- Chat Messages

-- name: CreateChatMessage :one
INSERT INTO cognitive.chat_messages (
    session_id,
    role,
    content,
    referenced_docs,
    tokens_used
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetChatMessagesBySession :many
SELECT * FROM cognitive.chat_messages
WHERE session_id = $1
ORDER BY created_at ASC;

-- name: GetRecentChatMessages :many
SELECT * FROM cognitive.chat_messages
WHERE session_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: CountChatMessagesBySession :one
SELECT COUNT(*) FROM cognitive.chat_messages
WHERE session_id = $1;

-- name: DeleteChatMessage :exec
DELETE FROM cognitive.chat_messages
WHERE id = $1;
