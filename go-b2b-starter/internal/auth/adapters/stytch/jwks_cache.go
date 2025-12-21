package stytch

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/moasq/go-b2b-starter/internal/logger"
	"github.com/moasq/go-b2b-starter/internal/redis"
)

const (
	// Redis cache keys for JWKS
	jwksCacheKeyPattern = "auth:stytch:jwks:key:%s" // Individual public key by kid
	jwksCacheTTL        = 24 * time.Hour            // 24-hour cache
)

// JWKSCache manages caching of JSON Web Key Sets from Stytch.
//
// It fetches JWKS from Stytch's endpoint and caches public keys in Redis.
// This enables local JWT verification without making Stytch API calls
// on every request (saving 300-500ms per request).
type JWKSCache struct {
	jwksURL    string
	redis      redis.Client
	logger     logger.Logger
	httpClient *http.Client
}

// JWKS represents the JSON Web Key Set structure from Stytch.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a single JSON Web Key.
type JWK struct {
	Kid string `json:"kid"` // Key ID
	Kty string `json:"kty"` // Key type (RSA)
	N   string `json:"n"`   // Modulus (base64url encoded)
	E   string `json:"e"`   // Exponent (base64url encoded)
	Alg string `json:"alg"` // Algorithm (RS256)
	Use string `json:"use"` // Public key use (sig)
}

// serializedPublicKey represents RSA public key components for Redis storage.
type serializedPublicKey struct {
	N string `json:"n"` // Modulus (base64url encoded)
	E string `json:"e"` // Exponent (base64url encoded)
}

func NewJWKSCache(jwksURL string, redisClient redis.Client, logger logger.Logger) *JWKSCache {
	return &JWKSCache{
		jwksURL: jwksURL,
		redis:   redisClient,
		logger:  logger,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetPublicKey retrieves a public key by kid from cache or fetches from Stytch.
func (c *JWKSCache) GetPublicKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	// Try to get from Redis cache first
	cacheKey := fmt.Sprintf(jwksCacheKeyPattern, kid)
	cached, err := c.redis.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		var serialized serializedPublicKey
		if err := json.Unmarshal([]byte(cached), &serialized); err == nil {
			key, err := c.deserializePublicKey(&serialized)
			if err == nil {
				c.logger.Debug("public key fetched from Redis cache", logger.Fields{
					"kid": kid,
				})
				return key, nil
			}
			c.logger.Warn("failed to deserialize cached public key", logger.Fields{
				"kid":   kid,
				"error": err.Error(),
			})
		}
	}

	// Cache miss - fetch JWKS from Stytch
	c.logger.Info("fetching JWKS from Stytch", logger.Fields{
		"jwks_url": c.jwksURL,
		"kid":      kid,
	})

	jwks, err := c.fetchJWKS(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	// Find the key with matching kid
	for _, jwk := range jwks.Keys {
		if jwk.Kid == kid {
			publicKey, err := c.jwkToPublicKey(&jwk)
			if err != nil {
				return nil, fmt.Errorf("failed to convert JWK to public key: %w", err)
			}

			// Cache the key
			c.cachePublicKey(ctx, kid, &jwk)

			c.logger.Info("public key fetched and cached", logger.Fields{
				"kid": kid,
			})

			return publicKey, nil
		}
	}

	// Log available keys for debugging
	availableKids := make([]string, 0, len(jwks.Keys))
	for _, jwk := range jwks.Keys {
		availableKids = append(availableKids, jwk.Kid)
	}
	c.logger.Error("key not found in JWKS", logger.Fields{
		"kid":            kid,
		"available_kids": availableKids,
	})

	return nil, fmt.Errorf("key with ID %s not found in JWKS", kid)
}

// fetchJWKS fetches the JWKS from Stytch's endpoint.
func (c *JWKSCache) fetchJWKS(ctx context.Context) (*JWKS, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.jwksURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWKS request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("JWKS HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS JSON: %w", err)
	}

	c.logger.Debug("successfully fetched JWKS", logger.Fields{
		"keys_count": len(jwks.Keys),
	})

	return &jwks, nil
}

// jwkToPublicKey converts a JWK to an RSA public key.
func (c *JWKSCache) jwkToPublicKey(jwk *JWK) (*rsa.PublicKey, error) {
	// Decode modulus (n)
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	// Decode exponent (e)
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert to RSA public key
	n := new(big.Int).SetBytes(nBytes)

	// Convert exponent bytes to int
	var e int
	for i := 0; i < len(eBytes); i++ {
		e = e<<8 + int(eBytes[i])
	}

	return &rsa.PublicKey{N: n, E: e}, nil
}

// cachePublicKey stores a public key in Redis.
func (c *JWKSCache) cachePublicKey(ctx context.Context, kid string, jwk *JWK) {
	serialized := &serializedPublicKey{N: jwk.N, E: jwk.E}

	data, err := json.Marshal(serialized)
	if err != nil {
		c.logger.Warn("failed to marshal public key for caching", logger.Fields{
			"kid":   kid,
			"error": err.Error(),
		})
		return
	}

	cacheKey := fmt.Sprintf(jwksCacheKeyPattern, kid)
	if err := c.redis.Set(ctx, cacheKey, string(data), jwksCacheTTL); err != nil {
		c.logger.Warn("failed to cache public key in Redis", logger.Fields{
			"kid":   kid,
			"error": err.Error(),
		})
	}
}

// deserializePublicKey converts serialized key components back to RSA public key.
func (c *JWKSCache) deserializePublicKey(serialized *serializedPublicKey) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(serialized.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cached modulus: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(serialized.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cached exponent: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	var e int
	for i := 0; i < len(eBytes); i++ {
		e = e<<8 + int(eBytes[i])
	}

	return &rsa.PublicKey{N: n, E: e}, nil
}
