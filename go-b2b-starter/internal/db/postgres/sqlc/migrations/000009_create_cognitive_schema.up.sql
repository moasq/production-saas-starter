-- Cognitive Agent schema for RAG and AI-powered features
CREATE SCHEMA IF NOT EXISTS cognitive;

-- Ensure pgvector extension is available
CREATE EXTENSION IF NOT EXISTS vector;

-- Document embeddings for RAG (vector search)
CREATE TABLE cognitive.document_embeddings (
    id SERIAL PRIMARY KEY,
    document_id INTEGER NOT NULL REFERENCES documents.documents(id) ON DELETE CASCADE,
    organization_id INTEGER NOT NULL REFERENCES organizations.organizations(id) ON DELETE CASCADE,
    embedding vector(1536) NOT NULL, -- OpenAI text-embedding-3-small dimension
    content_hash VARCHAR(64),
    content_preview TEXT,
    chunk_index INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(document_id, chunk_index)
);

-- IVFFlat index for fast vector similarity search using cosine distance
CREATE INDEX idx_doc_embeddings_vector ON cognitive.document_embeddings
USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

CREATE INDEX idx_doc_embeddings_organization ON cognitive.document_embeddings(organization_id);
CREATE INDEX idx_doc_embeddings_document ON cognitive.document_embeddings(document_id);
CREATE INDEX idx_doc_embeddings_content_hash ON cognitive.document_embeddings(content_hash);

-- Chat sessions for conversational AI
CREATE TABLE cognitive.chat_sessions (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations.organizations(id) ON DELETE CASCADE,
    account_id INTEGER NOT NULL,
    title VARCHAR(500),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_chat_sessions_organization ON cognitive.chat_sessions(organization_id);
CREATE INDEX idx_chat_sessions_account ON cognitive.chat_sessions(account_id);
CREATE INDEX idx_chat_sessions_created_at ON cognitive.chat_sessions(created_at DESC);

-- Chat messages within sessions
CREATE TABLE cognitive.chat_messages (
    id SERIAL PRIMARY KEY,
    session_id INTEGER NOT NULL REFERENCES cognitive.chat_sessions(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    referenced_docs INTEGER[],
    tokens_used INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_role CHECK (role IN ('user', 'assistant', 'system'))
);

CREATE INDEX idx_chat_messages_session ON cognitive.chat_messages(session_id);
CREATE INDEX idx_chat_messages_created_at ON cognitive.chat_messages(created_at);

-- Auto-update triggers for updated_at
CREATE OR REPLACE FUNCTION cognitive.update_embeddings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER doc_embeddings_updated_at
    BEFORE UPDATE ON cognitive.document_embeddings
    FOR EACH ROW
    EXECUTE FUNCTION cognitive.update_embeddings_updated_at();

CREATE OR REPLACE FUNCTION cognitive.update_sessions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER chat_sessions_updated_at
    BEFORE UPDATE ON cognitive.chat_sessions
    FOR EACH ROW
    EXECUTE FUNCTION cognitive.update_sessions_updated_at();

-- Comments for documentation
COMMENT ON TABLE cognitive.document_embeddings IS 'Vector embeddings for documents using OpenAI text-embedding-3-small (1536 dimensions)';
COMMENT ON COLUMN cognitive.document_embeddings.embedding IS 'Vector embedding for semantic similarity search';
COMMENT ON COLUMN cognitive.document_embeddings.chunk_index IS 'Index for chunked documents (0 for single-chunk docs)';
COMMENT ON TABLE cognitive.chat_sessions IS 'Conversational AI sessions for RAG-based chat';
COMMENT ON TABLE cognitive.chat_messages IS 'Messages within chat sessions with role (user/assistant/system)';
