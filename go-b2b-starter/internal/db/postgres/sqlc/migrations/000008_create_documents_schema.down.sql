-- Drop documents schema
DROP TRIGGER IF EXISTS documents_updated_at ON documents.documents;
DROP FUNCTION IF EXISTS documents.update_documents_updated_at();
DROP TABLE IF EXISTS documents.documents;
DROP SCHEMA IF EXISTS documents;
