package httperr

import (
	"net/http"
)

// IsNotFoundError checks if the error is a NotFoundError
func IsNotFoundError(err error) bool {
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.StatusCode == http.StatusNotFound
	}
	return false
}

// IsConflictError checks if the error is a ConflictError (e.g., duplicate username)
func IsConflictError(err error) bool {
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.StatusCode == http.StatusConflict
	}
	return false
}

// IsBadRequestError checks if the error is a BadRequestError
func IsBadRequestError(err error) bool {
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.StatusCode == http.StatusBadRequest
	}
	return false
}

// IsAuthenticationError checks if the error is an authentication error
func IsAuthenticationError(err error) bool {
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.StatusCode == http.StatusUnauthorized
	}
	return false
}

// IsAuthorizationError checks if the error is an authorization error
func IsAuthorizationError(err error) bool {
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.StatusCode == http.StatusForbidden
	}
	return false
}

// IsInternalServerError checks if the error is an internal server error
func IsInternalServerError(err error) bool {
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.StatusCode == http.StatusInternalServerError
	}
	return false
}

// GetErrorCode extracts the error code from an HTTPError, or returns empty string
func GetErrorCode(err error) string {
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.Code
	}
	return ""
}

// GetErrorMessage extracts the error message from an HTTPError, or returns the error's message
func GetErrorMessage(err error) string {
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.Message
	}
	if err != nil {
		return err.Error()
	}
	return ""
}
