-- name: CreateUser :one
INSERT INTO users.users (
    email,
    password_hash,
    full_name,
    status,
    role
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users.users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users.users WHERE email = $1;

-- name: UpdateUser :one
UPDATE users.users
SET
    full_name = COALESCE(sqlc.narg('full_name'), full_name),
    avatar_url = COALESCE(sqlc.narg('avatar_url'), avatar_url),
    status = COALESCE(sqlc.narg('status'), status),
    role = COALESCE(sqlc.narg('role'), role)
WHERE id = $1
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users.users
SET
    password_hash = $2,
    password_changed_at = NOW()
WHERE id = $1;

-- name: UpdateUserEmailVerified :exec
UPDATE users.users
SET
    email_verified = TRUE,
    status = 'active'
WHERE id = $1;

-- name: UpdateUserLastLogin :exec
UPDATE users.users
SET
    last_login_at = NOW(),
    last_login_ip = $2
WHERE id = $1;

-- name: IncrementFailedLoginAttempts :exec
UPDATE users.users
SET failed_login_attempts = failed_login_attempts + 1
WHERE id = $1;

-- name: ResetFailedLoginAttempts :exec
UPDATE users.users
SET
    failed_login_attempts = 0,
    locked_until = NULL
WHERE id = $1;

-- name: LockUserAccount :exec
UPDATE users.users
SET locked_until = $2
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users.users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users.users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users.users;

-- Refresh Token Queries

-- name: CreateRefreshToken :one
INSERT INTO users.refresh_tokens (
    user_id,
    token_hash,
    expires_at,
    device_info,
    ip_address,
    user_agent
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT * FROM users.refresh_tokens
WHERE token_hash = $1 AND expires_at > NOW();

-- name: RevokeRefreshToken :exec
UPDATE users.refresh_tokens
SET
    revoked = TRUE,
    revoked_at = NOW()
WHERE token_hash = $1;

-- name: RevokeAllUserRefreshTokens :exec
UPDATE users.refresh_tokens
SET
    revoked = TRUE,
    revoked_at = NOW()
WHERE user_id = $1 AND revoked = FALSE;

-- name: DeleteExpiredRefreshTokens :exec
DELETE FROM users.refresh_tokens
WHERE expires_at < NOW() OR (revoked = TRUE AND revoked_at < NOW() - INTERVAL '7 days');

-- Email Verification Token Queries

-- name: CreateEmailVerificationToken :one
INSERT INTO users.email_verification_tokens (
    user_id,
    token,
    expires_at
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetEmailVerificationTokenByToken :one
SELECT * FROM users.email_verification_tokens
WHERE token = $1 AND expires_at > NOW() AND used_at IS NULL;

-- name: MarkEmailVerificationTokenUsed :exec
UPDATE users.email_verification_tokens
SET used_at = NOW()
WHERE id = $1;

-- name: DeleteExpiredEmailVerificationTokens :exec
DELETE FROM users.email_verification_tokens
WHERE expires_at < NOW() OR used_at IS NOT NULL;

-- Password Reset Token Queries

-- name: CreatePasswordResetToken :one
INSERT INTO users.password_reset_tokens (
    user_id,
    token_hash,
    expires_at
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetPasswordResetTokenByHash :one
SELECT * FROM users.password_reset_tokens
WHERE token_hash = $1 AND expires_at > NOW() AND used_at IS NULL;

-- name: MarkPasswordResetTokenUsed :exec
UPDATE users.password_reset_tokens
SET used_at = NOW()
WHERE id = $1;

-- name: DeleteAllPasswordResetTokensForUser :exec
DELETE FROM users.password_reset_tokens WHERE user_id = $1;

-- name: DeleteExpiredPasswordResetTokens :exec
DELETE FROM users.password_reset_tokens
WHERE expires_at < NOW() OR used_at IS NOT NULL;
