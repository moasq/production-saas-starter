// Package stytch provides Stytch B2B authentication integration.
//
// This package implements the auth.AuthProvider interface using Stytch
// as the identity provider. It handles JWT verification, JWKS caching,
// and RBAC policy management.
//
// # Architecture
//
// The adapter uses a two-tier verification strategy:
//  1. Fast Path: Local JWT verification using cached JWKS (Redis)
//  2. Slow Path: Stytch API verification (fallback)
//
// This optimization saves 300-500ms per request for the common case.
//
// # Components
//
//   - StytchAuthAdapter: Main entry point implementing auth.AuthProvider
//   - TokenVerifier: JWT verification with local/API fallback
//   - JWKSCache: Public key caching in Redis
//   - RBACPolicyService: Role permission resolution
//
// # Usage
//
//	cfg, err := stytch.LoadConfig()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	adapter, err := stytch.NewStytchAuthAdapter(cfg, redisClient, logger)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Use as auth.AuthProvider
//	identity, err := adapter.VerifyToken(ctx, token)
package stytch

import (
	"context"
	"fmt"

	"github.com/moasq/go-b2b-starter/pkg/auth"
	"github.com/moasq/go-b2b-starter/pkg/logger"
	"github.com/moasq/go-b2b-starter/pkg/redis"
	"github.com/stytchauth/stytch-go/v16/stytch/b2b/b2bstytchapi"
)

// StytchAuthAdapter implements auth.AuthProvider using Stytch B2B.
//
// It provides authentication and authorization using Stytch's
// session management and RBAC capabilities.
type StytchAuthAdapter struct {
	client        *b2bstytchapi.API
	tokenVerifier *TokenVerifier
	policyService *RBACPolicyService
	cfg           *Config
	logger        logger.Logger
}

// Ensure StytchAuthAdapter implements auth.AuthProvider.
var _ auth.AuthProvider = (*StytchAuthAdapter)(nil)

// It initializes the Stytch client, JWKS cache, and RBAC policy service.
// Returns an error if configuration or client initialization fails.
func NewStytchAuthAdapter(
	cfg *Config,
	redisClient redis.Client,
	log logger.Logger,
) (*StytchAuthAdapter, error) {
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid stytch config: %w", err)
	}

	// Create Stytch API client
	client, err := b2bstytchapi.NewClient(cfg.ProjectID, cfg.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to create stytch client: %w", err)
	}

	// Create JWKS cache for local JWT verification
	jwksCache := NewJWKSCache(cfg.JWKSURL, redisClient, log)

	// Create RBAC policy service for permission resolution
	policyService := NewRBACPolicyService(client, redisClient, log)

	// Create token verifier with two-tier strategy
	tokenVerifier := NewTokenVerifier(client, jwksCache, policyService, cfg, log)

	return &StytchAuthAdapter{
		client:        client,
		tokenVerifier: tokenVerifier,
		policyService: policyService,
		cfg:           cfg,
		logger:        log,
	}, nil
}

// NewStytchAuthAdapterWithClient creates an adapter with an existing Stytch client.
//
// This is useful for testing or when you want to reuse an existing client.
func NewStytchAuthAdapterWithClient(
	client *b2bstytchapi.API,
	cfg *Config,
	redisClient redis.Client,
	log logger.Logger,
) *StytchAuthAdapter {
	jwksCache := NewJWKSCache(cfg.JWKSURL, redisClient, log)
	policyService := NewRBACPolicyService(client, redisClient, log)
	tokenVerifier := NewTokenVerifier(client, jwksCache, policyService, cfg, log)

	return &StytchAuthAdapter{
		client:        client,
		tokenVerifier: tokenVerifier,
		policyService: policyService,
		cfg:           cfg,
		logger:        log,
	}
}

// VerifyToken validates the supplied session JWT and returns an Identity.
//
// This implements auth.AuthProvider.VerifyToken.
//
// The verification uses a two-tier strategy:
//  1. Fast Path: Local JWT verification using cached JWKS
//  2. Slow Path: Stytch API verification (fallback)
//
// Returns auth.ErrInvalidToken if the token is invalid.
// Returns auth.ErrTokenExpired if the token has expired.
// Returns auth.ErrEmailNotVerified if email is not verified.
func (a *StytchAuthAdapter) VerifyToken(ctx context.Context, token string) (*auth.Identity, error) {
	if token == "" {
		return nil, auth.ErrInvalidToken
	}

	identity, err := a.tokenVerifier.Verify(ctx, token)
	if err != nil {
		a.logger.Debug("token verification failed", logger.Fields{
			"error": err.Error(),
		})
		return nil, err
	}

	a.logger.Debug("token verified successfully", logger.Fields{
		"user_id":         identity.UserID,
		"email":           identity.Email,
		"organization_id": identity.OrganizationID,
		"roles_count":     len(identity.Roles),
		"permissions_count": len(identity.Permissions),
	})

	return identity, nil
}

// Client returns the underlying Stytch API client.
//
// This is useful for advanced operations not covered by auth.AuthProvider,
// such as member management, organization settings, etc.
func (a *StytchAuthAdapter) Client() *b2bstytchapi.API {
	return a.client
}

// Config returns the Stytch configuration.
func (a *StytchAuthAdapter) Config() *Config {
	return a.cfg
}

// PolicyService returns the RBAC policy service.
//
// This is useful for permission queries outside the normal auth flow.
func (a *StytchAuthAdapter) PolicyService() *RBACPolicyService {
	return a.policyService
}
