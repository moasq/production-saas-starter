package stytch

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// JWTParser provides JWT token parsing utilities.
//
// This parser can decode JWT tokens without verifying the signature,
// which is useful for extracting the key ID (kid) from the header
// before signature verification.
type JWTParser struct{}

func NewJWTParser() *JWTParser {
	return &JWTParser{}
}

// ParseWithoutVerification decodes a JWT token without verifying the signature.
//
// This extracts the header and claims from the token for inspection.
// The signature is NOT verified - use this only for extracting metadata
// like the key ID (kid) before performing proper verification.
//
// Returns:
//   - header: JWT header containing algorithm (alg), key ID (kid), etc.
//   - claims: JWT claims containing user information
//   - err: Error if the token format is invalid
func (p *JWTParser) ParseWithoutVerification(token string) (header map[string]any, claims map[string]any, err error) {
	// JWT format: header.payload.signature
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, nil, fmt.Errorf("invalid JWT format: expected 3 parts, got %d", len(parts))
	}

	// Decode header (first part)
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode JWT header: %w", err)
	}

	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, nil, fmt.Errorf("failed to parse JWT header JSON: %w", err)
	}

	// Decode payload (second part)
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, nil, fmt.Errorf("failed to parse JWT claims JSON: %w", err)
	}

	return header, claims, nil
}

// ExtractKeyID extracts the key ID (kid) from a JWT token header.
//
// This is a convenience method for getting the kid without needing
// to handle the full header map.
func (p *JWTParser) ExtractKeyID(token string) (string, error) {
	header, _, err := p.ParseWithoutVerification(token)
	if err != nil {
		return "", err
	}

	kid, ok := header["kid"].(string)
	if !ok {
		return "", fmt.Errorf("kid not found in JWT header")
	}

	return kid, nil
}
