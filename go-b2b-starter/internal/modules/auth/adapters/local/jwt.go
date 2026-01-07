package local

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenType represents the type of token.
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// Claims represents the JWT claims for local auth.
type Claims struct {
	jwt.RegisteredClaims
	UserID        int32  `json:"user_id"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Role          string `json:"role"`
	TokenType     string `json:"token_type"`
}

// TokenPair contains both access and refresh tokens.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// JWTManager handles JWT token operations.
type JWTManager struct {
	config *Config
}

// NewJWTManager creates a new JWT manager.
func NewJWTManager(config *Config) *JWTManager {
	return &JWTManager{config: config}
}

// GenerateTokenPair generates both access and refresh tokens.
func (m *JWTManager) GenerateTokenPair(userID int32, email string, emailVerified bool, role string) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(m.config.AccessTokenDuration)
	refreshExpiry := now.Add(m.config.RefreshTokenDuration)

	// Generate access token
	accessClaims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.config.Issuer,
			Audience:  jwt.ClaimStrings{m.config.Audience},
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			ID:        uuid.New().String(),
		},
		UserID:        userID,
		Email:         email,
		EmailVerified: emailVerified,
		Role:          role,
		TokenType:     string(TokenTypeAccess),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(m.config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshClaims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.config.Issuer,
			Audience:  jwt.ClaimStrings{m.config.Audience},
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(refreshExpiry),
			ID:        uuid.New().String(),
		},
		UserID:        userID,
		Email:         email,
		EmailVerified: emailVerified,
		Role:          role,
		TokenType:     string(TokenTypeRefresh),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(m.config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    accessExpiry,
		TokenType:    "Bearer",
	}, nil
}

// VerifyAccessToken verifies an access token and returns the claims.
func (m *JWTManager) VerifyAccessToken(tokenString string) (*Claims, error) {
	return m.verifyToken(tokenString, TokenTypeAccess)
}

// VerifyRefreshToken verifies a refresh token and returns the claims.
func (m *JWTManager) VerifyRefreshToken(tokenString string) (*Claims, error) {
	return m.verifyToken(tokenString, TokenTypeRefresh)
}

func (m *JWTManager) verifyToken(tokenString string, expectedType TokenType) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.config.PublicKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims.TokenType != string(expectedType) {
		return nil, ErrInvalidToken
	}

	// Verify issuer and audience
	if claims.Issuer != m.config.Issuer {
		return nil, ErrInvalidToken
	}

	if !claims.VerifyAudience(m.config.Audience, true) {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// HashToken creates a SHA-256 hash of a token for storage.
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// Errors
var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token has expired")
)
