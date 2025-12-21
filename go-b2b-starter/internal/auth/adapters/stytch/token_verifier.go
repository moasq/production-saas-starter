package stytch

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/moasq/go-b2b-starter/internal/auth"
	"github.com/moasq/go-b2b-starter/internal/logger"
	"github.com/stytchauth/stytch-go/v16/stytch/b2b/b2bstytchapi"
	"github.com/stytchauth/stytch-go/v16/stytch/b2b/sessions"
	"github.com/stytchauth/stytch-go/v16/stytch/stytcherror"
)

// TokenVerifier verifies Stytch session JWTs using a two-tier strategy:
//  1. Fast Path: Local JWT verification using cached JWKS (no API calls)
//  2. Slow Path: Stytch API verification (fallback when local fails)
//
// This optimization saves 300-500ms per request for the common case.
type TokenVerifier struct {
	client        *b2bstytchapi.API
	jwksCache     *JWKSCache
	jwtParser     *JWTParser
	policyService *RBACPolicyService
	cfg           *Config
	logger        logger.Logger
}

func NewTokenVerifier(
	client *b2bstytchapi.API,
	jwksCache *JWKSCache,
	policyService *RBACPolicyService,
	cfg *Config,
	logger logger.Logger,
) *TokenVerifier {
	return &TokenVerifier{
		client:        client,
		jwksCache:     jwksCache,
		jwtParser:     NewJWTParser(),
		policyService: policyService,
		cfg:           cfg,
		logger:        logger,
	}
}

// internalClaims holds parsed JWT claims before conversion to auth.Identity.
type internalClaims struct {
	Subject        string
	Email          string
	EmailVerified  bool
	OrganizationID string
	Roles          []string
	Permissions    []auth.Permission
	IssuedAt       time.Time
	ExpiresAt      time.Time
	NotBefore      time.Time
	Issuer         string
	Audience       []string
	Raw            map[string]any
}

// Verify validates the token and returns an Identity.
//
// It tries local verification first (fast path), falling back to
// Stytch API verification if local verification fails.
func (v *TokenVerifier) Verify(ctx context.Context, token string) (*auth.Identity, error) {
	// Check for test mode (DANGEROUS - only for development)
	if v.cfg.DisableSessionVerification {
		v.logger.Warn("session verification disabled - test mode only", logger.Fields{})
		return v.verifyWithoutSignature(ctx, token)
	}

	// Fast path: Local JWT verification
	identity, err := v.verifyLocally(ctx, token)
	if err == nil {
		v.logger.Debug("token verified locally (fast path)", logger.Fields{
			"user_id": identity.UserID,
			"email":   identity.Email,
		})
		return identity, nil
	}

	// Log fast path failure
	v.logger.Warn("local verification failed, trying Stytch API", logger.Fields{
		"error": err.Error(),
	})

	// Slow path: Stytch API verification
	return v.verifyViaAPI(ctx, token)
}

// verifyLocally verifies the token using cached JWKS (fast path).
func (v *TokenVerifier) verifyLocally(ctx context.Context, token string) (*auth.Identity, error) {
	// 1. Parse token header to get key ID
	kid, err := v.jwtParser.ExtractKeyID(token)
	if err != nil {
		return nil, fmt.Errorf("failed to extract key ID: %w", err)
	}

	// 2. Get public key from cache
	publicKey, err := v.jwksCache.GetPublicKey(ctx, kid)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	// 3. Verify token signature
	jwtToken, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil {
		if strings.Contains(err.Error(), "expired") {
			return nil, auth.ErrTokenExpired
		}
		return nil, auth.ErrInvalidToken
	}

	if !jwtToken.Valid {
		return nil, auth.ErrInvalidToken
	}

	// 4. Parse claims from token
	_, claimsMap, err := v.jwtParser.ParseWithoutVerification(token)
	if err != nil {
		return nil, auth.ErrInvalidToken
	}

	claims := v.parseClaimsFromMap(claimsMap)

	// 5. Validate claims
	if err := v.validateClaims(claims); err != nil {
		return nil, err
	}

	// 6. Derive permissions from roles
	permissions := v.derivePermissions(ctx, claims.Roles)

	// 7. Convert to Identity
	return &auth.Identity{
		UserID:         claims.Subject,
		Email:          claims.Email,
		EmailVerified:  claims.EmailVerified,
		OrganizationID: claims.OrganizationID,
		Roles:          v.convertRoles(claims.Roles),
		Permissions:    permissions,
		ExpiresAt:      claims.ExpiresAt,
		Raw:            claims.Raw,
	}, nil
}

// verifyViaAPI verifies the token using Stytch API (slow path).
func (v *TokenVerifier) verifyViaAPI(ctx context.Context, token string) (*auth.Identity, error) {
	// Add timeout
	ctx, cancel := context.WithTimeout(ctx, v.cfg.APITimeout)
	defer cancel()

	req := &sessions.AuthenticateParams{
		SessionJWT: token,
	}
	if v.cfg.SessionDurationMinutes > 0 {
		req.SessionDurationMinutes = v.cfg.SessionDurationMinutes
	}

	resp, err := v.client.Sessions.Authenticate(ctx, req)
	if err != nil {
		return nil, v.translateStytchError(err)
	}

	member := resp.Member
	session := resp.MemberSession

	// Check email verification
	if !member.EmailAddressVerified {
		return nil, auth.ErrEmailNotVerified
	}

	// Derive permissions from roles
	permissions := v.derivePermissions(ctx, session.Roles)

	// Build identity
	identity := &auth.Identity{
		UserID:         session.MemberID,
		Email:          member.EmailAddress,
		EmailVerified:  member.EmailAddressVerified,
		OrganizationID: session.OrganizationID,
		Roles:          v.convertRoles(session.Roles),
		Permissions:    permissions,
		ExpiresAt:      timeValue(session.ExpiresAt),
		Raw: map[string]any{
			"member_session": session,
			"member":         member,
			"custom_claims":  session.CustomClaims,
		},
	}

	v.logger.Debug("token verified via Stytch API (slow path)", logger.Fields{
		"user_id": identity.UserID,
		"email":   identity.Email,
	})

	return identity, nil
}

// verifyWithoutSignature parses the token without verifying signature (test mode only).
func (v *TokenVerifier) verifyWithoutSignature(ctx context.Context, token string) (*auth.Identity, error) {
	_, claimsMap, err := v.jwtParser.ParseWithoutVerification(token)
	if err != nil {
		return nil, auth.ErrInvalidToken
	}

	claims := v.parseClaimsFromMap(claimsMap)
	permissions := v.derivePermissions(ctx, claims.Roles)

	return &auth.Identity{
		UserID:         claims.Subject,
		Email:          claims.Email,
		EmailVerified:  claims.EmailVerified,
		OrganizationID: claims.OrganizationID,
		Roles:          v.convertRoles(claims.Roles),
		Permissions:    permissions,
		ExpiresAt:      claims.ExpiresAt,
		Raw:            claims.Raw,
	}, nil
}

// parseClaimsFromMap extracts claims from JWT payload.
func (v *TokenVerifier) parseClaimsFromMap(claimsMap map[string]any) *internalClaims {
	claims := &internalClaims{
		Raw: claimsMap,
	}

	// Extract subject
	if sub, ok := claimsMap["sub"].(string); ok {
		claims.Subject = sub
	}

	// Extract email from Stytch session authentication factors
	// Format: https://stytch.com/session.authentication_factors[].email_factor.email_address
	if sessionObj, ok := claimsMap["https://stytch.com/session"].(map[string]any); ok {
		if factors, ok := sessionObj["authentication_factors"].([]any); ok {
			for _, factor := range factors {
				if factorMap, ok := factor.(map[string]any); ok {
					if emailFactor, ok := factorMap["email_factor"].(map[string]any); ok {
						if emailAddr, ok := emailFactor["email_address"].(string); ok {
							claims.Email = emailAddr
							break
						}
					}
				}
			}
		}

		// Extract roles from session
		if rolesIface, ok := sessionObj["roles"].([]any); ok {
			roles := make([]string, 0, len(rolesIface))
			for _, r := range rolesIface {
				if roleStr, ok := r.(string); ok {
					roles = append(roles, roleStr)
				}
			}
			claims.Roles = roles
		}
	}

	// Fallback to standard email claim
	if claims.Email == "" {
		if email, ok := claimsMap["email"].(string); ok {
			claims.Email = email
		}
	}

	// Extract email_verified
	if verified, ok := claimsMap["email_verified"].(bool); ok {
		claims.EmailVerified = verified
	} else if verified, ok := claimsMap["https://stytch.com/email_verified"].(bool); ok {
		claims.EmailVerified = verified
	}

	// Extract organization ID from Stytch custom claim
	// Format: https://stytch.com/organization.organization_id
	if orgObj, ok := claimsMap["https://stytch.com/organization"].(map[string]any); ok {
		if orgID, ok := orgObj["organization_id"].(string); ok {
			claims.OrganizationID = orgID
		}
	}
	// Fallback to standard claims
	if claims.OrganizationID == "" {
		if orgID, ok := claimsMap["organization_id"].(string); ok {
			claims.OrganizationID = orgID
		} else if orgID, ok := claimsMap["org_id"].(string); ok {
			claims.OrganizationID = orgID
		}
	}

	// Parse timestamps
	claims.IssuedAt = parseNumericTime(claimsMap["iat"])
	claims.ExpiresAt = parseNumericTime(claimsMap["exp"])
	claims.NotBefore = parseNumericTime(claimsMap["nbf"])

	// Parse issuer
	if iss, ok := claimsMap["iss"].(string); ok {
		claims.Issuer = iss
	}

	// Parse audience
	claims.Audience = parseStringSlice(claimsMap["aud"])

	// Fallback for roles
	if len(claims.Roles) == 0 {
		claims.Roles = parseStringSlice(claimsMap["roles"])
	}

	return claims
}

// validateClaims validates security-critical claims.
func (v *TokenVerifier) validateClaims(claims *internalClaims) error {
	now := time.Now()

	// Check expiry
	if !claims.ExpiresAt.IsZero() && now.After(claims.ExpiresAt) {
		return auth.ErrTokenExpired
	}

	// Check not before
	if !claims.NotBefore.IsZero() && now.Before(claims.NotBefore) {
		return auth.ErrInvalidToken
	}

	// Validate issuer (must be from Stytch)
	if claims.Issuer != "" && !strings.Contains(strings.ToLower(claims.Issuer), "stytch.com") {
		return auth.ErrIssuerMismatch
	}

	// Check email verification if claim is present
	if _, hasEmailVerified := claims.Raw["email_verified"]; hasEmailVerified {
		if !claims.EmailVerified {
			return auth.ErrEmailNotVerified
		}
	}

	return nil
}

// derivePermissions derives permissions from roles.
//
// Fast path: Use hardcoded permissions for standard roles (no API calls).
// Slow path: Fetch from Stytch RBAC policy for custom roles.
func (v *TokenVerifier) derivePermissions(ctx context.Context, roles []string) []auth.Permission {
	permSet := make(map[auth.Permission]struct{})

	for _, roleStr := range roles {
		if roleStr == "" {
			continue
		}

		// Normalize role
		normalizedRole := auth.NormalizeRole(roleStr)

		// Fast path: Use hardcoded permissions for standard roles
		if perms := auth.GetRolePermissions(normalizedRole); len(perms) > 0 {
			for _, p := range perms {
				permSet[p] = struct{}{}
			}
			continue
		}

		// Slow path: Fetch from Stytch RBAC policy
		if v.policyService != nil {
			stytchPerms, err := v.policyService.GetRolePermissions(ctx, roleStr)
			if err != nil {
				v.logger.Warn("failed to get role permissions from Stytch", logger.Fields{
					"role":  roleStr,
					"error": err.Error(),
				})
				continue
			}

			for _, p := range stytchPerms {
				permSet[p] = struct{}{}
			}
		}
	}

	// Convert set to slice
	if len(permSet) == 0 {
		return nil
	}

	permissions := make([]auth.Permission, 0, len(permSet))
	for p := range permSet {
		permissions = append(permissions, p)
	}

	return permissions
}

// convertRoles converts string role names to auth.Role.
func (v *TokenVerifier) convertRoles(roles []string) []auth.Role {
	if len(roles) == 0 {
		return nil
	}

	result := make([]auth.Role, 0, len(roles))
	seen := make(map[auth.Role]struct{})

	for _, roleStr := range roles {
		role := auth.NormalizeRole(roleStr)
		if _, exists := seen[role]; !exists {
			seen[role] = struct{}{}
			result = append(result, role)
		}
	}

	return result
}

// translateStytchError converts Stytch errors to auth errors.
func (v *TokenVerifier) translateStytchError(err error) error {
	var stErr *stytcherror.Error
	if errors.As(err, &stErr) {
		if strings.Contains(strings.ToLower(string(stErr.ErrorType)), "expired") {
			return auth.ErrTokenExpired
		}
		switch stErr.StatusCode {
		case 401, 403, 404:
			return auth.ErrInvalidToken
		default:
			if stErr.StatusCode >= 500 {
				return fmt.Errorf("stytch service error: %w", err)
			}
			return auth.ErrInvalidToken
		}
	}
	return fmt.Errorf("stytch error: %w", err)
}

// Helper functions

func parseStringSlice(value any) []string {
	switch v := value.(type) {
	case string:
		if v == "" {
			return nil
		}
		return []string{v}
	case []string:
		return v
	case []any:
		res := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				res = append(res, s)
			}
		}
		return res
	default:
		return nil
	}
}

func parseNumericTime(value any) time.Time {
	switch v := value.(type) {
	case float64:
		return time.Unix(int64(v), 0)
	case int64:
		return time.Unix(v, 0)
	case int:
		return time.Unix(int64(v), 0)
	default:
		return time.Time{}
	}
}

func timeValue(ts *time.Time) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.UTC()
}
