-- Enable pgvector extension for vector similarity search
CREATE EXTENSION IF NOT EXISTS vector;

-- Resource embeddings table
-- Stores vector embeddings generated from resource text content for semantic similarity search
CREATE TABLE resource_embeddings (
    id SERIAL PRIMARY KEY,
    resource_id INTEGER NOT NULL REFERENCES public.example_resources(id) ON DELETE CASCADE,
    embedding vector(1536) NOT NULL, -- OpenAI text-embedding-3-small dimension is 1536
    organization_id INTEGER NOT NULL REFERENCES organizations.organizations(id) ON DELETE CASCADE,
    content_hash VARCHAR(64), -- SHA-256 hash for exact duplicate detection
    content_preview TEXT, -- First 500 chars of content for debugging
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(resource_id, organization_id) -- One embedding per resource per organization
);

-- Create ivfflat index for fast vector similarity search using cosine distance
-- ivfflat divides vectors into lists for approximate nearest neighbor search
-- lists=100 is a good starting point (adjust based on dataset size)
CREATE INDEX idx_resource_embeddings_vector
ON resource_embeddings
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

-- Regular indexes for lookups
CREATE INDEX idx_resource_embeddings_organization ON resource_embeddings(organization_id);
CREATE INDEX idx_resource_embeddings_content_hash ON resource_embeddings(content_hash);
CREATE INDEX idx_resource_embeddings_resource ON resource_embeddings(resource_id);

-- Duplicate candidates table
-- Stores potential duplicates found through vector similarity search and LLM adjudication
CREATE TABLE duplicate_candidates (
    id SERIAL PRIMARY KEY,
    resource_id INTEGER NOT NULL REFERENCES public.example_resources(id) ON DELETE CASCADE,
    candidate_resource_id INTEGER NOT NULL REFERENCES public.example_resources(id) ON DELETE CASCADE,
    similarity_score DECIMAL(5,4) NOT NULL, -- Vector cosine similarity score (0.0000 to 1.0000)
    detection_method VARCHAR(50) NOT NULL, -- 'exact_match' or 'llm_adjudicated'
    confidence_level VARCHAR(50), -- 'very_high', 'high', 'medium', 'low' (from LLM)
    llm_reason TEXT, -- LLM's explanation for duplicate decision
    llm_similar_fields JSONB, -- Fields identified as similar by LLM
    llm_response JSONB, -- Full LLM response for audit trail
    organization_id INTEGER NOT NULL REFERENCES organizations.organizations(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'confirmed', 'dismissed'
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CHECK (resource_id != candidate_resource_id) -- Prevent self-duplicates
);

-- Indexes for duplicate candidates
CREATE INDEX idx_duplicate_candidates_resource ON duplicate_candidates(resource_id);
CREATE INDEX idx_duplicate_candidates_candidate ON duplicate_candidates(candidate_resource_id);
CREATE INDEX idx_duplicate_candidates_organization ON duplicate_candidates(organization_id);
CREATE INDEX idx_duplicate_candidates_status ON duplicate_candidates(status);
CREATE INDEX idx_duplicate_candidates_method ON duplicate_candidates(detection_method);

-- Composite index for common query pattern: find all duplicates for a resource
CREATE INDEX idx_duplicate_candidates_resource_org ON duplicate_candidates(resource_id, organization_id);

-- Auto-update trigger for updated_at
CREATE OR REPLACE FUNCTION update_resource_embeddings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER resource_embeddings_updated_at
    BEFORE UPDATE ON resource_embeddings
    FOR EACH ROW
    EXECUTE FUNCTION update_resource_embeddings_updated_at();

CREATE OR REPLACE FUNCTION update_duplicate_candidates_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER duplicate_candidates_updated_at
    BEFORE UPDATE ON duplicate_candidates
    FOR EACH ROW
    EXECUTE FUNCTION update_duplicate_candidates_updated_at();

-- Comments for documentation
COMMENT ON TABLE resource_embeddings IS 'Stores vector embeddings for resources using OpenAI text-embedding-3-small (1536 dimensions)';
COMMENT ON COLUMN resource_embeddings.embedding IS 'Vector embedding for semantic similarity search (1536 dimensions from OpenAI)';
COMMENT ON COLUMN resource_embeddings.content_hash IS 'SHA-256 hash of normalized content for exact duplicate detection';
COMMENT ON INDEX idx_resource_embeddings_vector IS 'IVFFlat index for fast approximate nearest neighbor search using cosine distance';

COMMENT ON TABLE duplicate_candidates IS 'Stores potential duplicate resources found via vector similarity and LLM adjudication';
COMMENT ON COLUMN duplicate_candidates.similarity_score IS 'Cosine similarity score from pgvector (0.0000 = completely different, 1.0000 = identical)';
COMMENT ON COLUMN duplicate_candidates.detection_method IS 'How the duplicate was detected: exact_match (similarity >= 0.95) or llm_adjudicated (similarity >= 0.85, confirmed by LLM)';
