BEGIN;

-- Drop indexes
DROP INDEX IF EXISTS file_manager.idx_file_assets_entity;
DROP INDEX IF EXISTS file_manager.idx_file_assets_category;  
DROP INDEX IF EXISTS file_manager.idx_file_assets_context;
DROP INDEX IF EXISTS file_manager.idx_file_assets_created_at;

-- Drop tables in reverse order
DROP TABLE IF EXISTS file_manager.file_assets;
DROP TABLE IF EXISTS file_manager.file_contexts;
DROP TABLE IF EXISTS file_manager.file_categories;

-- Drop schema
DROP SCHEMA IF EXISTS file_manager CASCADE;

COMMIT;