// Package domain defines the user domain models and interfaces.
package domain

import "time"

// UserStatus represents the status of a user account.
type UserStatus string

const (
	UserStatusPendingVerification UserStatus = "pending_verification"
	UserStatusActive              UserStatus = "active"
	UserStatusSuspended           UserStatus = "suspended"
	UserStatusDeleted             UserStatus = "deleted"
)

// UserRole represents the role of a user.
type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

// User represents a user in the system.
type User struct {
	ID                  int32      `json:"id"`
	Email               string     `json:"email"`
	PasswordHash        string     `json:"-"`
	EmailVerified       bool       `json:"email_verified"`
	FullName            *string    `json:"full_name,omitempty"`
	AvatarURL           *string    `json:"avatar_url,omitempty"`
	Status              UserStatus `json:"status"`
	Role                UserRole   `json:"role"`
	FailedLoginAttempts int32      `json:"-"`
	LockedUntil         *time.Time `json:"-"`
	PasswordChangedAt   *time.Time `json:"-"`
	LastLoginAt         *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP         *string    `json:"-"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// IsLocked returns true if the user account is currently locked.
func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.LockedUntil)
}

// RefreshToken represents a stored refresh token.
type RefreshToken struct {
	ID         int32      `json:"id"`
	UserID     int32      `json:"user_id"`
	TokenHash  string     `json:"-"`
	ExpiresAt  time.Time  `json:"expires_at"`
	Revoked    bool       `json:"revoked"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	DeviceInfo *string    `json:"device_info,omitempty"`
	IPAddress  *string    `json:"ip_address,omitempty"`
	UserAgent  *string    `json:"user_agent,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// EmailVerificationToken represents an email verification token.
type EmailVerificationToken struct {
	ID        int32     `json:"id"`
	UserID    int32     `json:"user_id"`
	Token     string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// PasswordResetToken represents a password reset token.
type PasswordResetToken struct {
	ID        int32      `json:"id"`
	UserID    int32      `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
