CREATE SCHEMA IF NOT EXISTS file_manager;

-- Create file categories table (similar to asset_types)
CREATE TABLE file_manager.file_categories (
    id SMALLINT PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,  -- 'document', 'image', 'archive'
    max_size_bytes BIGINT NOT NULL
);

-- Create file contexts table 
CREATE TABLE file_manager.file_contexts (
    id SMALLINT PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL  -- 'invoice', 'receipt', 'contract', etc.
);

-- Create file_assets table (following assets pattern)
CREATE TABLE file_manager.file_assets (
    id SERIAL PRIMARY KEY,
    file_name VARCHAR(255) NOT NULL,
    original_file_name VARCHAR(255) NOT NULL,
    storage_path VARCHAR(1000) NOT NULL, 
    bucket_name VARCHAR(50) NOT NULL,
    file_size BIGINT NOT NULL CHECK (file_size > 0),
    mime_type VARCHAR(100) NOT NULL,
    file_category_id SMALLINT NOT NULL REFERENCES file_manager.file_categories(id),
    file_context_id SMALLINT NOT NULL REFERENCES file_manager.file_contexts(id),
    is_public BOOLEAN DEFAULT false,
    entity_type VARCHAR(50),  -- 'user', 'invoice', 'contract', etc.
    entity_id INTEGER,        -- The ID of the related entity
    purpose VARCHAR(100),     -- Additional purpose description
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes following the same pattern
CREATE INDEX idx_file_assets_entity ON file_manager.file_assets(entity_type, entity_id);
CREATE INDEX idx_file_assets_category ON file_manager.file_assets(file_category_id);
CREATE INDEX idx_file_assets_context ON file_manager.file_assets(file_context_id);
CREATE INDEX idx_file_assets_created_at ON file_manager.file_assets(created_at DESC);

-- Insert default categories
INSERT INTO file_manager.file_categories (id, name, max_size_bytes) VALUES
(1, 'document', 52428800),  -- 50MB
(2, 'image', 10485760),     -- 10MB  
(3, 'archive', 104857600);  -- 100MB

-- Insert default contexts
INSERT INTO file_manager.file_contexts (id, name) VALUES
(1, 'invoice'),
(2, 'receipt'),
(3, 'contract'),
(4, 'report'),
(5, 'profile'),
(6, 'general');