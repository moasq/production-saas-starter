-- Drop users schema
DROP TRIGGER IF EXISTS update_users_updated_at ON users.users;
DROP FUNCTION IF EXISTS users.update_updated_at_column();
DROP TABLE IF EXISTS users.password_reset_tokens;
DROP TABLE IF EXISTS users.email_verification_tokens;
DROP TABLE IF EXISTS users.refresh_tokens;
DROP TABLE IF EXISTS users.users;
DROP TYPE IF EXISTS users.user_role;
DROP TYPE IF EXISTS users.user_status;
DROP SCHEMA IF EXISTS users;
