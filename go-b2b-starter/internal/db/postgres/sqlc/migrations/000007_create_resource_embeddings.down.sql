-- Drop triggers
DROP TRIGGER IF EXISTS duplicate_candidates_updated_at ON duplicate_candidates;
DROP TRIGGER IF EXISTS resource_embeddings_updated_at ON resource_embeddings;

-- Drop trigger functions
DROP FUNCTION IF EXISTS update_duplicate_candidates_updated_at();
DROP FUNCTION IF EXISTS update_resource_embeddings_updated_at();

-- Drop tables (cascade to remove dependent objects)
DROP TABLE IF EXISTS duplicate_candidates CASCADE;
DROP TABLE IF EXISTS resource_embeddings CASCADE;

-- Note: We don't drop the vector extension as it might be used by other tables
-- If you want to remove it completely, uncomment the line below:
-- DROP EXTENSION IF EXISTS vector CASCADE;
