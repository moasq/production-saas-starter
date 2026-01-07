package local

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/moasq/go-b2b-starter/internal/modules/auth"
	"github.com/moasq/go-b2b-starter/internal/modules/users/domain"
	"github.com/moasq/go-b2b-starter/internal/platform/logger"
)

// Adapter implements auth.AuthProvider for self-hosted JWT authentication.
type Adapter struct {
	config         *Config
	jwtManager     *JWTManager
	passwordHasher *PasswordHasher
	userRepo       domain.UserRepository
	refreshRepo    domain.RefreshTokenRepository
	log            logger.Logger
}

// NewAdapter creates a new local auth adapter.
func NewAdapter(
	config *Config,
	userRepo domain.UserRepository,
	refreshRepo domain.RefreshTokenRepository,
	log logger.Logger,
) *Adapter {
	return &Adapter{
		config:         config,
		jwtManager:     NewJWTManager(config),
		passwordHasher: NewPasswordHasher(config),
		userRepo:       userRepo,
		refreshRepo:    refreshRepo,
		log:            log,
	}
}

// VerifyToken implements auth.AuthProvider.
func (a *Adapter) VerifyToken(ctx context.Context, token string) (*auth.Identity, error) {
	claims, err := a.jwtManager.VerifyAccessToken(token)
	if err != nil {
		return nil, err
	}

	// Map role to permissions
	role := auth.Role(claims.Role)
	permissions := a.getRolePermissions(role)

	return &auth.Identity{
		UserID:        strconv.Itoa(int(claims.UserID)),
		Email:         claims.Email,
		EmailVerified: claims.EmailVerified,
		Roles:         []auth.Role{role},
		Permissions:   permissions,
		ExpiresAt:     claims.ExpiresAt.Time,
		Raw: map[string]any{
			"user_id":   claims.UserID,
			"token_id":  claims.ID,
			"issued_at": claims.IssuedAt.Time,
		},
	}, nil
}

// Register creates a new user account.
func (a *Adapter) Register(ctx context.Context, email, password, fullName string) (*domain.User, error) {
	// Validate password
	if err := a.passwordHasher.ValidatePassword(password); err != nil {
		return nil, err
	}

	// Hash password
	passwordHash, err := a.passwordHasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	var fullNamePtr *string
	if fullName != "" {
		fullNamePtr = &fullName
	}

	user, err := a.userRepo.Create(ctx, domain.CreateUserInput{
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullNamePtr,
		Role:         domain.RoleUser,
	})
	if err != nil {
		return nil, err
	}

	a.log.Info("user registered", map[string]any{
		"user_id": user.ID,
		"email":   user.Email,
	})

	return user, nil
}

// Login authenticates a user and returns tokens.
func (a *Adapter) Login(ctx context.Context, email, password, ipAddress, userAgent string) (*TokenPair, *domain.User, error) {
	// Get user by email
	user, err := a.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, nil, domain.ErrInvalidCredentials
	}

	// Check if account is locked
	if user.IsLocked() {
		return nil, nil, domain.ErrAccountLocked
	}

	// Check if account is active
	if user.Status != domain.UserStatusActive && user.Status != domain.UserStatusPendingVerification {
		return nil, nil, domain.ErrAccountInactive
	}

	// Verify password
	valid, err := a.passwordHasher.Verify(password, user.PasswordHash)
	if err != nil || !valid {
		// Increment failed attempts
		if incrErr := a.userRepo.IncrementFailedLoginAttempts(ctx, user.ID); incrErr != nil {
			a.log.Error("failed to increment failed login attempts", map[string]any{
				"error":   incrErr.Error(),
				"user_id": user.ID,
			})
		}

		// Check if should lock
		user.FailedLoginAttempts++
		if user.FailedLoginAttempts >= int32(a.config.MaxFailedAttempts) {
			lockUntil := time.Now().Add(a.config.LockoutDuration)
			if lockErr := a.userRepo.LockAccount(ctx, user.ID, lockUntil); lockErr != nil {
				a.log.Error("failed to lock account", map[string]any{
					"error":   lockErr.Error(),
					"user_id": user.ID,
				})
			}
		}

		return nil, nil, domain.ErrInvalidCredentials
	}

	// Reset failed attempts on successful login
	if user.FailedLoginAttempts > 0 {
		if resetErr := a.userRepo.ResetFailedLoginAttempts(ctx, user.ID); resetErr != nil {
			a.log.Error("failed to reset failed login attempts", map[string]any{
				"error":   resetErr.Error(),
				"user_id": user.ID,
			})
		}
	}

	// Generate tokens
	tokenPair, err := a.jwtManager.GenerateTokenPair(user.ID, user.Email, user.EmailVerified, string(user.Role))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Store refresh token hash
	refreshTokenHash := HashToken(tokenPair.RefreshToken)
	expiresAt := time.Now().Add(a.config.RefreshTokenDuration)

	var ipPtr, uaPtr *string
	if ipAddress != "" {
		ipPtr = &ipAddress
	}
	if userAgent != "" {
		uaPtr = &userAgent
	}

	if _, err := a.refreshRepo.Create(ctx, user.ID, refreshTokenHash, expiresAt, nil, ipPtr, uaPtr); err != nil {
		a.log.Error("failed to store refresh token", map[string]any{
			"error":   err.Error(),
			"user_id": user.ID,
		})
	}

	// Update last login
	if updateErr := a.userRepo.UpdateLastLogin(ctx, user.ID, ipAddress); updateErr != nil {
		a.log.Error("failed to update last login", map[string]any{
			"error":   updateErr.Error(),
			"user_id": user.ID,
		})
	}

	a.log.Info("user logged in", map[string]any{
		"user_id": user.ID,
		"email":   user.Email,
	})

	return tokenPair, user, nil
}

// RefreshTokens generates new tokens using a refresh token.
func (a *Adapter) RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// Verify refresh token
	claims, err := a.jwtManager.VerifyRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Check if refresh token exists and is not revoked
	tokenHash := HashToken(refreshToken)
	storedToken, err := a.refreshRepo.GetByHash(ctx, tokenHash)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if storedToken.Revoked {
		// Token reuse detected - revoke all tokens for user
		if revokeErr := a.refreshRepo.RevokeAllForUser(ctx, claims.UserID); revokeErr != nil {
			a.log.Error("failed to revoke all tokens after reuse detection", map[string]any{
				"error":   revokeErr.Error(),
				"user_id": claims.UserID,
			})
		}
		return nil, ErrInvalidToken
	}

	// Revoke the old refresh token
	if err := a.refreshRepo.Revoke(ctx, tokenHash); err != nil {
		a.log.Error("failed to revoke old refresh token", map[string]any{
			"error":   err.Error(),
			"user_id": claims.UserID,
		})
	}

	// Generate new token pair
	tokenPair, err := a.jwtManager.GenerateTokenPair(claims.UserID, claims.Email, claims.EmailVerified, claims.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	// Store new refresh token
	newTokenHash := HashToken(tokenPair.RefreshToken)
	expiresAt := time.Now().Add(a.config.RefreshTokenDuration)

	if _, err := a.refreshRepo.Create(ctx, claims.UserID, newTokenHash, expiresAt, nil, nil, nil); err != nil {
		a.log.Error("failed to store new refresh token", map[string]any{
			"error":   err.Error(),
			"user_id": claims.UserID,
		})
	}

	return tokenPair, nil
}

// Logout revokes a refresh token.
func (a *Adapter) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := HashToken(refreshToken)
	return a.refreshRepo.Revoke(ctx, tokenHash)
}

// LogoutAll revokes all refresh tokens for a user.
func (a *Adapter) LogoutAll(ctx context.Context, userID int32) error {
	return a.refreshRepo.RevokeAllForUser(ctx, userID)
}

// ChangePassword changes a user's password.
func (a *Adapter) ChangePassword(ctx context.Context, userID int32, currentPassword, newPassword string) error {
	// Get user
	user, err := a.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify current password
	valid, err := a.passwordHasher.Verify(currentPassword, user.PasswordHash)
	if err != nil || !valid {
		return domain.ErrInvalidCredentials
	}

	// Validate new password
	if err := a.passwordHasher.ValidatePassword(newPassword); err != nil {
		return err
	}

	// Hash new password
	newPasswordHash, err := a.passwordHasher.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := a.userRepo.UpdatePassword(ctx, userID, newPasswordHash); err != nil {
		return err
	}

	// Revoke all refresh tokens
	if err := a.refreshRepo.RevokeAllForUser(ctx, userID); err != nil {
		a.log.Error("failed to revoke tokens after password change", map[string]any{
			"error":   err.Error(),
			"user_id": userID,
		})
	}

	a.log.Info("user changed password", map[string]any{
		"user_id": userID,
	})

	return nil
}

// getRolePermissions returns permissions for a role.
func (a *Adapter) getRolePermissions(role auth.Role) []auth.Permission {
	switch role {
	case auth.RoleAdmin:
		return []auth.Permission{
			auth.NewPermission("users", "read"),
			auth.NewPermission("users", "write"),
			auth.NewPermission("users", "delete"),
			auth.NewPermission("files", "read"),
			auth.NewPermission("files", "write"),
			auth.NewPermission("files", "delete"),
			auth.NewPermission("documents", "read"),
			auth.NewPermission("documents", "write"),
			auth.NewPermission("documents", "delete"),
			auth.NewPermission("admin", "read"),
			auth.NewPermission("admin", "write"),
		}
	case auth.RoleUser:
		return []auth.Permission{
			auth.NewPermission("files", "read"),
			auth.NewPermission("files", "write"),
			auth.NewPermission("documents", "read"),
			auth.NewPermission("documents", "write"),
		}
	default:
		return []auth.Permission{}
	}
}
