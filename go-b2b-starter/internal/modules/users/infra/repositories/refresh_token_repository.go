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

// RefreshTokenRepository implements domain.RefreshTokenRepository using SQLC.
type RefreshTokenRepository struct {
	queries *sqlc.Queries
}

// NewRefreshTokenRepository creates a new refresh token repository.
func NewRefreshTokenRepository(queries *sqlc.Queries) *RefreshTokenRepository {
	return &RefreshTokenRepository{queries: queries}
}

// Create creates a new refresh token.
func (r *RefreshTokenRepository) Create(ctx context.Context, userID int32, tokenHash string, expiresAt time.Time, deviceInfo, ipAddress, userAgent *string) (*domain.RefreshToken, error) {
	var device, ip, ua sql.NullString

	if deviceInfo != nil {
		device = sql.NullString{String: *deviceInfo, Valid: true}
	}
	if ipAddress != nil {
		ip = sql.NullString{String: *ipAddress, Valid: true}
	}
	if userAgent != nil {
		ua = sql.NullString{String: *userAgent, Valid: true}
	}

	token, err := r.queries.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		UserID:     userID,
		TokenHash:  tokenHash,
		ExpiresAt:  expiresAt,
		DeviceInfo: device,
		IpAddress:  ip,
		UserAgent:  ua,
	})
	if err != nil {
		return nil, err
	}

	return mapSQLCRefreshTokenToDomain(token), nil
}

// GetByHash retrieves a valid refresh token by its hash.
func (r *RefreshTokenRepository) GetByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	token, err := r.queries.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidToken
		}
		return nil, err
	}
	return mapSQLCRefreshTokenToDomain(token), nil
}

// Revoke revokes a refresh token.
func (r *RefreshTokenRepository) Revoke(ctx context.Context, tokenHash string) error {
	return r.queries.RevokeRefreshToken(ctx, tokenHash)
}

// RevokeAllForUser revokes all refresh tokens for a user.
func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID int32) error {
	return r.queries.RevokeAllUserRefreshTokens(ctx, userID)
}

// DeleteExpired deletes expired and revoked tokens.
func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.queries.DeleteExpiredRefreshTokens(ctx)
}

// mapSQLCRefreshTokenToDomain converts SQLC refresh token to domain refresh token.
func mapSQLCRefreshTokenToDomain(t sqlc.UsersRefreshToken) *domain.RefreshToken {
	token := &domain.RefreshToken{
		ID:        t.ID,
		UserID:    t.UserID,
		TokenHash: t.TokenHash,
		ExpiresAt: t.ExpiresAt,
		Revoked:   t.Revoked,
		CreatedAt: t.CreatedAt,
	}

	if t.DeviceInfo.Valid {
		token.DeviceInfo = &t.DeviceInfo.String
	}
	if t.IpAddress.Valid {
		token.IPAddress = &t.IpAddress.String
	}
	if t.UserAgent.Valid {
		token.UserAgent = &t.UserAgent.String
	}
	if t.RevokedAt.Valid {
		token.RevokedAt = &t.RevokedAt.Time
	}

	return token
}
