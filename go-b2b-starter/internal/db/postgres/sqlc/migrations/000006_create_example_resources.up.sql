-- Create example_resources table
-- This table demonstrates the full architecture pattern with:
-- - File attachments
-- - OCR/LLM processing
-- - Multi-status workflow
-- - RBAC integration
-- - Multi-tenancy
-- - Approval workflow
-- - Audit tracking

CREATE TABLE IF NOT EXISTS example_resources (
    id SERIAL PRIMARY KEY,
    resource_number VARCHAR(100) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,

    -- Status workflow
    status_id SMALLINT NOT NULL DEFAULT 1,

    -- File attachment
    file_id INTEGER REFERENCES file_manager.file_assets(id) ON DELETE SET NULL,

    -- AI Processing results
    extracted_data JSONB DEFAULT '{}',  -- OCR output
    processed_data JSONB DEFAULT '{}',  -- LLM structured output
    confidence DECIMAL(5,4),  -- AI confidence score (0.0000 to 1.0000)

    -- Multi-tenancy (required)
    organization_id INTEGER NOT NULL REFERENCES organizations.organizations(id) ON DELETE CASCADE,
    created_by_account_id INTEGER REFERENCES organizations.accounts(id) ON DELETE SET NULL,

    -- Approval workflow
    approval_status VARCHAR(50) DEFAULT 'pending',
    approval_assigned_to_id INTEGER REFERENCES organizations.accounts(id) ON DELETE SET NULL,
    approval_action_taker_id INTEGER REFERENCES organizations.accounts(id) ON DELETE SET NULL,
    approval_notes TEXT,

    -- Additional metadata
    metadata JSONB DEFAULT '{}',

    -- Audit fields
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_example_resources_organization ON example_resources(organization_id);
CREATE INDEX idx_example_resources_status ON example_resources(status_id);
CREATE INDEX idx_example_resources_created_by ON example_resources(created_by_account_id);
CREATE INDEX idx_example_resources_file ON example_resources(file_id);
CREATE INDEX idx_example_resources_approval_assigned ON example_resources(approval_assigned_to_id);
CREATE INDEX idx_example_resources_created_at ON example_resources(created_at DESC);
CREATE INDEX idx_example_resources_active ON example_resources(is_active);

-- Full text search on title and description
CREATE INDEX idx_example_resources_search ON example_resources USING gin(to_tsvector('english', coalesce(title, '') || ' ' || coalesce(description, '')));

-- Trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_example_resources_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_example_resources_updated_at
    BEFORE UPDATE ON example_resources
    FOR EACH ROW
    EXECUTE FUNCTION update_example_resources_updated_at();

-- Comments for documentation
COMMENT ON TABLE example_resources IS 'Example module demonstrating Clean Architecture patterns with file uploads, OCR/LLM processing, RBAC, approval workflows, and multi-tenancy';
COMMENT ON COLUMN example_resources.extracted_data IS 'Raw OCR-extracted text and metadata';
COMMENT ON COLUMN example_resources.processed_data IS 'LLM-processed structured data';
COMMENT ON COLUMN example_resources.confidence IS 'AI confidence score between 0 and 1';
COMMENT ON COLUMN example_resources.approval_status IS 'Workflow status: pending, approved, rejected';
