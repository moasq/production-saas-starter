package polar

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

// VerifyWebhookSignature verifies that a webhook request came from Polar
// by validating the HMAC-SHA256 signature using the Standard Webhooks specification
//
// Polar.sh uses the Standard Webhooks format (same as Svix) where the signed content is:
// {webhook-id}.{webhook-timestamp}.{body}
//
// Parameters:
//   - secret: The webhook secret from Polar Dashboard
//   - webhookID: The Webhook-Id header value
//   - timestamp: The Webhook-Timestamp header value
//   - payload: The raw request body (must be the exact bytes received)
//   - signature: The signature from the Webhook-Signature header
//
// Returns:
//   - error if verification fails, nil if successful
func VerifyWebhookSignature(secret string, webhookID string, timestamp string, payload []byte, signature string) error {
	if secret == "" {
		return fmt.Errorf("webhook secret is not configured")
	}

	if signature == "" {
		return fmt.Errorf("webhook signature is missing from request")
	}

	if webhookID == "" {
		return fmt.Errorf("webhook ID is missing from request")
	}

	if timestamp == "" {
		return fmt.Errorf("webhook timestamp is missing from request")
	}

	// Strip version prefix (e.g., "v1,") from Polar's signature
	if strings.Contains(signature, ",") {
		parts := strings.Split(signature, ",")
		if len(parts) == 2 {
			signature = parts[1] // Get the actual signature after "v1,"
		}
	}

	// Decode base64 signature to bytes (Polar sends base64-encoded HMAC)
	signatureBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	// Construct the signed content according to Standard Webhooks spec
	// Format: {webhook-id}.{webhook-timestamp}.{body}
	signedContent := webhookID + "." + timestamp + "." + string(payload)

	// Compute HMAC-SHA256 of the signed content
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedContent))
	expectedSignatureBytes := mac.Sum(nil)

	// Use constant-time comparison to prevent timing attacks
	if !hmac.Equal(signatureBytes, expectedSignatureBytes) {
		return fmt.Errorf("webhook signature verification failed: signature mismatch")
	}

	return nil
}

// ComputeWebhookSignature computes the HMAC-SHA256 signature for a payload
// using the Standard Webhooks format
// This is useful for testing webhook signature verification
// Returns the signature in base64 format to match Polar's format
func ComputeWebhookSignature(secret string, webhookID string, timestamp string, payload []byte) string {
	// Construct signed content: {webhook-id}.{webhook-timestamp}.{body}
	signedContent := webhookID + "." + timestamp + "." + string(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedContent))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
