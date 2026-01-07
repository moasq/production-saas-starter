-- Users schema for self-hosted authentication
-- This replaces B2B organizations/accounts with B2C users

-- Create users schema
CREATE SCHEMA IF NOT EXISTS users;

-- User status enum
CREATE TYPE users.user_status AS ENUM (
    'pending_verification',
    'active',
    'suspended',
    'deleted'
);

-- User role enum
CREATE TYPE users.user_role AS ENUM (
    'user',
    'admin'
);

-- Users table
CREATE TABLE users.users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    full_name VARCHAR(255),
    avatar_url VARCHAR(500),
    status users.user_status NOT NULL DEFAULT 'pending_verification',
    role users.user_role NOT NULL DEFAULT 'user',
    failed_login_attempts INTEGER NOT NULL DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    password_changed_at TIMESTAMP WITH TIME ZONE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    last_login_ip VARCHAR(45),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Refresh tokens table (for JWT token rotation)
CREATE TABLE users.refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    revoked_at TIMESTAMP WITH TIME ZONE,
    device_info VARCHAR(255),
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Email verification tokens
CREATE TABLE users.email_verification_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
    token VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Password reset tokens
CREATE TABLE users.password_reset_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_users_email ON users.users(email);
CREATE INDEX idx_users_status ON users.users(status);
CREATE INDEX idx_refresh_tokens_user_id ON users.refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON users.refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires_at ON users.refresh_tokens(expires_at);
CREATE INDEX idx_email_verification_tokens_user_id ON users.email_verification_tokens(user_id);
CREATE INDEX idx_email_verification_tokens_token ON users.email_verification_tokens(token);
CREATE INDEX idx_password_reset_tokens_user_id ON users.password_reset_tokens(user_id);
CREATE INDEX idx_password_reset_tokens_token_hash ON users.password_reset_tokens(token_hash);

-- Updated at trigger function
CREATE OR REPLACE FUNCTION users.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to users table
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users.users
    FOR EACH ROW
    EXECUTE FUNCTION users.update_updated_at_column();
