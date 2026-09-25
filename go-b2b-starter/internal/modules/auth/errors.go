package auth

import "errors"

var ErrUnauthorized = errors.New("authentication required")
