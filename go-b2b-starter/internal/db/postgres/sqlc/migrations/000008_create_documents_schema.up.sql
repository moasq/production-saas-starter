-- Documents schema for PDF upload and text extraction
CREATE SCHEMA IF NOT EXISTS documents;

-- Documents table
CREATE TABLE documents.documents (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL REFERENCES organizations.organizations(id) ON DELETE CASCADE,
    file_asset_id INTEGER NOT NULL REFERENCES file_manager.file_assets(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    file_name VARCHAR(500) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    extracted_text TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_status CHECK (status IN ('pending', 'processing', 'processed', 'failed'))
);

-- Indexes
CREATE INDEX idx_documents_organization ON documents.documents(organization_id);
CREATE INDEX idx_documents_file_asset ON documents.documents(file_asset_id);
CREATE INDEX idx_documents_status ON documents.documents(status);
CREATE INDEX idx_documents_created_at ON documents.documents(created_at DESC);

-- Auto-update trigger for updated_at
CREATE OR REPLACE FUNCTION documents.update_documents_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER documents_updated_at
    BEFORE UPDATE ON documents.documents
    FOR EACH ROW
    EXECUTE FUNCTION documents.update_documents_updated_at();

-- Comments for documentation
COMMENT ON TABLE documents.documents IS 'Stores uploaded documents (PDFs) with extracted text for RAG';
COMMENT ON COLUMN documents.documents.extracted_text IS 'Text extracted from PDF using OCR or direct parsing';
COMMENT ON COLUMN documents.documents.status IS 'Processing status: pending, processing, processed, failed';
