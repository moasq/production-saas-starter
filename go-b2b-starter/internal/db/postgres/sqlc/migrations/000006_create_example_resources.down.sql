-- Rollback example_resources table

-- Drop trigger first
DROP TRIGGER IF EXISTS trigger_update_example_resources_updated_at ON example_resources;
DROP FUNCTION IF EXISTS update_example_resources_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_example_resources_search;
DROP INDEX IF EXISTS idx_example_resources_active;
DROP INDEX IF EXISTS idx_example_resources_created_at;
DROP INDEX IF EXISTS idx_example_resources_approval_assigned;
DROP INDEX IF EXISTS idx_example_resources_file;
DROP INDEX IF EXISTS idx_example_resources_created_by;
DROP INDEX IF EXISTS idx_example_resources_status;
DROP INDEX IF EXISTS idx_example_resources_organization;

-- Drop table
DROP TABLE IF EXISTS example_resources;
