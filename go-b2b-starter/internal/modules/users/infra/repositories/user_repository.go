// Package repositories provides database implementations for user domain interfaces.
package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/moasq/go-b2b-starter/internal/db/postgres/sqlc"
	"github.com/moasq/go-b2b-starter/internal/modules/users/domain"
)

// UserRepository implements domain.UserRepository using SQLC.
type UserRepository struct {
	queries *sqlc.Queries
}

// NewUserRepository creates a new user repository.
func NewUserRepository(queries *sqlc.Queries) *UserRepository {
	return &UserRepository{queries: queries}
}

// Create creates a new user.
func (r *UserRepository) Create(ctx context.Context, input domain.CreateUserInput) (*domain.User, error) {
	var fullName sql.NullString
	if input.FullName != nil {
		fullName = sql.NullString{String: *input.FullName, Valid: true}
	}

	user, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        input.Email,
		PasswordHash: input.PasswordHash,
		FullName:     fullName,
		Status:       string(domain.UserStatusPendingVerification),
		Role:         string(input.Role),
	})
	if err != nil {
		return nil, err
	}

	return mapSQLCUserToDomain(user), nil
}

// GetByID retrieves a user by ID.
func (r *UserRepository) GetByID(ctx context.Context, id int32) (*domain.User, error) {
	user, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return mapSQLCUserToDomain(user), nil
}

// GetByEmail retrieves a user by email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return mapSQLCUserToDomain(user), nil
}

// Update updates a user.
func (r *UserRepository) Update(ctx context.Context, id int32, input domain.UpdateUserInput) (*domain.User, error) {
	var fullName, avatarURL, status, role sql.NullString

	if input.FullName != nil {
		fullName = sql.NullString{String: *input.FullName, Valid: true}
	}
	if input.AvatarURL != nil {
		avatarURL = sql.NullString{String: *input.AvatarURL, Valid: true}
	}
	if input.Status != nil {
		status = sql.NullString{String: string(*input.Status), Valid: true}
	}
	if input.Role != nil {
		role = sql.NullString{String: string(*input.Role), Valid: true}
	}

	user, err := r.queries.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:        id,
		FullName:  fullName,
		AvatarUrl: avatarURL,
		Status:    status,
		Role:      role,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return mapSQLCUserToDomain(user), nil
}

// UpdatePassword updates a user's password.
func (r *UserRepository) UpdatePassword(ctx context.Context, id int32, passwordHash string) error {
	return r.queries.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		ID:           id,
		PasswordHash: passwordHash,
	})
}

// UpdateEmailVerified marks a user's email as verified.
func (r *UserRepository) UpdateEmailVerified(ctx context.Context, id int32) error {
	return r.queries.UpdateUserEmailVerified(ctx, id)
}

// UpdateLastLogin updates the last login timestamp and IP.
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id int32, ip string) error {
	return r.queries.UpdateUserLastLogin(ctx, sqlc.UpdateUserLastLoginParams{
		ID:          id,
		LastLoginIp: sql.NullString{String: ip, Valid: ip != ""},
	})
}

// IncrementFailedLoginAttempts increments the failed login counter.
func (r *UserRepository) IncrementFailedLoginAttempts(ctx context.Context, id int32) error {
	return r.queries.IncrementFailedLoginAttempts(ctx, id)
}

// ResetFailedLoginAttempts resets the failed login counter.
func (r *UserRepository) ResetFailedLoginAttempts(ctx context.Context, id int32) error {
	return r.queries.ResetFailedLoginAttempts(ctx, id)
}

// LockAccount locks a user account until the specified time.
func (r *UserRepository) LockAccount(ctx context.Context, id int32, until time.Time) error {
	return r.queries.LockUserAccount(ctx, sqlc.LockUserAccountParams{
		ID:          id,
		LockedUntil: sql.NullTime{Time: until, Valid: true},
	})
}

// Delete deletes a user.
func (r *UserRepository) Delete(ctx context.Context, id int32) error {
	return r.queries.DeleteUser(ctx, id)
}

// List lists users with pagination.
func (r *UserRepository) List(ctx context.Context, limit, offset int32) ([]*domain.User, error) {
	users, err := r.queries.ListUsers(ctx, sqlc.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*domain.User, len(users))
	for i, u := range users {
		result[i] = mapSQLCUserToDomain(u)
	}
	return result, nil
}

// Count returns the total number of users.
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	return r.queries.CountUsers(ctx)
}

// mapSQLCUserToDomain converts SQLC user to domain user.
func mapSQLCUserToDomain(u sqlc.UsersUser) *domain.User {
	user := &domain.User{
		ID:                  u.ID,
		Email:               u.Email,
		PasswordHash:        u.PasswordHash,
		EmailVerified:       u.EmailVerified,
		Status:              domain.UserStatus(u.Status),
		Role:                domain.UserRole(u.Role),
		FailedLoginAttempts: u.FailedLoginAttempts,
		CreatedAt:           u.CreatedAt,
		UpdatedAt:           u.UpdatedAt,
	}

	if u.FullName.Valid {
		user.FullName = &u.FullName.String
	}
	if u.AvatarUrl.Valid {
		user.AvatarURL = &u.AvatarUrl.String
	}
	if u.LockedUntil.Valid {
		user.LockedUntil = &u.LockedUntil.Time
	}
	if u.PasswordChangedAt.Valid {
		user.PasswordChangedAt = &u.PasswordChangedAt.Time
	}
	if u.LastLoginAt.Valid {
		user.LastLoginAt = &u.LastLoginAt.Time
	}
	if u.LastLoginIp.Valid {
		user.LastLoginIP = &u.LastLoginIp.String
	}

	return user
}
