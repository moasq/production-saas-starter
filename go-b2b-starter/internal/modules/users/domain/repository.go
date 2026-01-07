package domain

import (
	"context"
	"time"
)

// CreateUserInput contains the data needed to create a new user.
type CreateUserInput struct {
	Email        string
	PasswordHash string
	FullName     *string
	Role         UserRole
}

// UpdateUserInput contains the optional fields that can be updated.
type UpdateUserInput struct {
	FullName  *string
	AvatarURL *string
	Status    *UserStatus
	Role      *UserRole
}

// UserRepository defines the interface for user data access.
type UserRepository interface {
	// Create creates a new user.
	Create(ctx context.Context, input CreateUserInput) (*User, error)

	// GetByID retrieves a user by ID.
	GetByID(ctx context.Context, id int32) (*User, error)

	// GetByEmail retrieves a user by email.
	GetByEmail(ctx context.Context, email string) (*User, error)

	// Update updates a user's profile.
	Update(ctx context.Context, id int32, input UpdateUserInput) (*User, error)

	// UpdatePassword updates a user's password hash.
	UpdatePassword(ctx context.Context, id int32, passwordHash string) error

	// UpdateEmailVerified marks a user's email as verified.
	UpdateEmailVerified(ctx context.Context, id int32) error

	// UpdateLastLogin updates the last login timestamp and IP.
	UpdateLastLogin(ctx context.Context, id int32, ip string) error

	// IncrementFailedLoginAttempts increments the failed login counter.
	IncrementFailedLoginAttempts(ctx context.Context, id int32) error

	// ResetFailedLoginAttempts resets the failed login counter.
	ResetFailedLoginAttempts(ctx context.Context, id int32) error

	// LockAccount locks a user account until the specified time.
	LockAccount(ctx context.Context, id int32, until time.Time) error

	// Delete deletes a user.
	Delete(ctx context.Context, id int32) error

	// List lists users with pagination.
	List(ctx context.Context, limit, offset int32) ([]*User, error)

	// Count returns the total number of users.
	Count(ctx context.Context) (int64, error)
}

// RefreshTokenRepository defines the interface for refresh token data access.
type RefreshTokenRepository interface {
	// Create creates a new refresh token.
	Create(ctx context.Context, userID int32, tokenHash string, expiresAt time.Time, deviceInfo, ipAddress, userAgent *string) (*RefreshToken, error)

	// GetByHash retrieves a valid refresh token by its hash.
	GetByHash(ctx context.Context, tokenHash string) (*RefreshToken, error)

	// Revoke revokes a refresh token.
	Revoke(ctx context.Context, tokenHash string) error

	// RevokeAllForUser revokes all refresh tokens for a user.
	RevokeAllForUser(ctx context.Context, userID int32) error

	// DeleteExpired deletes expired and revoked tokens.
	DeleteExpired(ctx context.Context) error
}

// EmailVerificationTokenRepository defines the interface for email verification tokens.
type EmailVerificationTokenRepository interface {
	// Create creates a new email verification token.
	Create(ctx context.Context, userID int32, token string, expiresAt time.Time) (*EmailVerificationToken, error)

	// GetByToken retrieves a token by its value.
	GetByToken(ctx context.Context, token string) (*EmailVerificationToken, error)

	// MarkUsed marks a token as used.
	MarkUsed(ctx context.Context, id int32) error

	// DeleteExpired deletes expired tokens.
	DeleteExpired(ctx context.Context) error
}

// PasswordResetTokenRepository defines the interface for password reset tokens.
type PasswordResetTokenRepository interface {
	// Create creates a new password reset token.
	Create(ctx context.Context, userID int32, tokenHash string, expiresAt time.Time) (*PasswordResetToken, error)

	// GetByHash retrieves a token by its hash.
	GetByHash(ctx context.Context, tokenHash string) (*PasswordResetToken, error)

	// MarkUsed marks a token as used.
	MarkUsed(ctx context.Context, id int32) error

	// DeleteAllForUser deletes all tokens for a user.
	DeleteAllForUser(ctx context.Context, userID int32) error

	// DeleteExpired deletes expired tokens.
	DeleteExpired(ctx context.Context) error
}
