-- Drop cognitive schema
DROP TRIGGER IF EXISTS chat_sessions_updated_at ON cognitive.chat_sessions;
DROP TRIGGER IF EXISTS doc_embeddings_updated_at ON cognitive.document_embeddings;
DROP FUNCTION IF EXISTS cognitive.update_sessions_updated_at();
DROP FUNCTION IF EXISTS cognitive.update_embeddings_updated_at();
DROP TABLE IF EXISTS cognitive.chat_messages;
DROP TABLE IF EXISTS cognitive.chat_sessions;
DROP TABLE IF EXISTS cognitive.document_embeddings;
DROP SCHEMA IF EXISTS cognitive;
