package auth

import "errors"

// Authentication and authorization errors.
//
// These errors are returned by the auth package and can be checked
// by application code to handle specific error cases.
var (
	// ErrUnauthorized is returned when authentication is required but not provided.
	// HTTP status: 401 Unauthorized
	ErrUnauthorized = errors.New("authentication required")

	// ErrInvalidToken is returned when the provided token is malformed or invalid.
	// HTTP status: 401 Unauthorized
	ErrInvalidToken = errors.New("invalid token")

	// ErrTokenExpired is returned when the token has expired.
	// HTTP status: 401 Unauthorized
	ErrTokenExpired = errors.New("token expired")

	// ErrEmailNotVerified is returned when the user's email is not verified.
	// HTTP status: 403 Forbidden
	ErrEmailNotVerified = errors.New("email not verified")

	// ErrForbidden is returned when the user lacks required permissions.
	// HTTP status: 403 Forbidden
	ErrForbidden = errors.New("insufficient permissions")

	// ErrOrganizationNotFound is returned when the organization cannot be found.
	// This typically means the organization in the token doesn't exist in our database.
	// HTTP status: 403 Forbidden
	ErrOrganizationNotFound = errors.New("organization not found")

	// ErrAccountNotFound is returned when the user's account cannot be found.
	// This typically means the user exists in the auth provider but not in our database.
	// HTTP status: 403 Forbidden
	ErrAccountNotFound = errors.New("account not found")

	// ErrMissingOrganization is returned when the token doesn't contain an organization ID.
	// HTTP status: 403 Forbidden
	ErrMissingOrganization = errors.New("no organization in token")

	// ErrMissingEmail is returned when the token doesn't contain an email.
	// HTTP status: 403 Forbidden
	ErrMissingEmail = errors.New("no email in token")

	// ErrAudienceMismatch is returned when the token audience doesn't match.
	// HTTP status: 401 Unauthorized
	ErrAudienceMismatch = errors.New("token audience mismatch")

	// ErrIssuerMismatch is returned when the token issuer doesn't match.
	// HTTP status: 401 Unauthorized
	ErrIssuerMismatch = errors.New("token issuer mismatch")
)

// IsAuthError returns true if the error is an authentication error (401).
func IsAuthError(err error) bool {
	return errors.Is(err, ErrUnauthorized) ||
		errors.Is(err, ErrInvalidToken) ||
		errors.Is(err, ErrTokenExpired) ||
		errors.Is(err, ErrAudienceMismatch) ||
		errors.Is(err, ErrIssuerMismatch)
}

// IsForbiddenError returns true if the error is an authorization error (403).
func IsForbiddenError(err error) bool {
	return errors.Is(err, ErrForbidden) ||
		errors.Is(err, ErrEmailNotVerified) ||
		errors.Is(err, ErrOrganizationNotFound) ||
		errors.Is(err, ErrAccountNotFound) ||
		errors.Is(err, ErrMissingOrganization) ||
		errors.Is(err, ErrMissingEmail)
}

// HTTPStatusCode returns the appropriate HTTP status code for an auth error.
//
// Returns:
//   - 401 for authentication errors
//   - 403 for authorization errors
//   - 500 for unknown errors
func HTTPStatusCode(err error) int {
	if IsAuthError(err) {
		return 401
	}
	if IsForbiddenError(err) {
		return 403
	}
	return 500
}
