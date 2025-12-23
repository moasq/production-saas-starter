package stytch

import (
	"errors"
	"fmt"

	"github.com/stytchauth/stytch-go/v16/stytch/stytcherror"
)

var (
	// ErrUnauthorized mirrors HTTP 401 responses.
	ErrUnauthorized = errors.New("stytch: unauthorized")
	// ErrForbidden mirrors HTTP 403 responses.
	ErrForbidden = errors.New("stytch: forbidden")
	// ErrNotFound mirrors HTTP 404 responses.
	ErrNotFound = errors.New("stytch: resource not found")
	// ErrConflict mirrors HTTP 409 responses.
	ErrConflict = errors.New("stytch: conflict")
	// ErrRateLimited mirrors HTTP 429 responses.
	ErrRateLimited = errors.New("stytch: rate limit exceeded")
	// ErrBadRequest mirrors HTTP 400 responses.
	ErrBadRequest = errors.New("stytch: bad request")
	// ErrInternal mirrors HTTP 5xx responses.
	ErrInternal = errors.New("stytch: internal server error")
	// ErrInvalidConfig surfaces configuration validation issues.
	ErrInvalidConfig = errors.New("stytch: invalid configuration")
	// ErrDuplicateSlug indicates organization slug already exists.
	ErrDuplicateSlug = errors.New("stytch: organization slug already exists")
)

// IsDuplicateSlugError checks if the error is a duplicate organization slug error
func IsDuplicateSlugError(err error) bool {
	if err == nil {
		return false
	}

	var stErr *stytcherror.Error
	if errors.As(err, &stErr) {
		return string(stErr.ErrorType) == "organization_slug_already_used"
	}

	return false
}

// MapError inspects a returned error and maps it to one of the sentinel errors above.
func MapError(err error) error {
	if err == nil {
		return nil
	}

	var stErr *stytcherror.Error
	if errors.As(err, &stErr) {
		// Check specific error types first
		if string(stErr.ErrorType) == "organization_slug_already_used" {
			return fmt.Errorf("%w: %s", ErrDuplicateSlug, stErr.Error())
		}

		// Then check status codes
		switch stErr.StatusCode {
		case 400:
			return fmt.Errorf("%w: %s", ErrBadRequest, stErr.Error())
		case 401:
			return fmt.Errorf("%w: %s", ErrUnauthorized, stErr.Error())
		case 403:
			return fmt.Errorf("%w: %s", ErrForbidden, stErr.Error())
		case 404:
			return fmt.Errorf("%w: %s", ErrNotFound, stErr.Error())
		case 409:
			return fmt.Errorf("%w: %s", ErrConflict, stErr.Error())
		case 429:
			return fmt.Errorf("%w: %s", ErrRateLimited, stErr.Error())
		default:
			if stErr.StatusCode >= 500 {
				return fmt.Errorf("%w: %s", ErrInternal, stErr.Error())
			}
			return fmt.Errorf("stytch: unexpected status %d: %w", stErr.StatusCode, stErr)
		}
	}

	return err
}
