package postgres

import (
	"context"
	"errors"
	"log"
	"time"
)

const (
	maxRetries    = 3
	retryDelay    = 100 * time.Millisecond
	maxRetryDelay = 1 * time.Second
)

// RetryOperation executes a database operation with exponential backoff retry
func RetryOperation(ctx context.Context, operation func(context.Context) error) error {
	var err error

	backoff := retryDelay
	for i := 0; i < maxRetries; i++ {
		// Execute the operation
		err = operation(ctx)

		// If no error or context cancelled, return immediately
		if err == nil || errors.Is(err, context.Canceled) {
			return err
		}

		// Log the error for debugging
		log.Printf("Database operation failed (attempt %d/%d): %v", i+1, maxRetries, err)

		// Don't retry if the final attempt
		if i == maxRetries-1 {
			break
		}

		// Wait with backoff before retrying
		select {
		case <-time.After(backoff):
			// Exponential backoff
			backoff *= 2
			if backoff > maxRetryDelay {
				backoff = maxRetryDelay
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return err
}

// CreateDBContext creates a context with an appropriate timeout for database operations
func CreateDBContext(parent context.Context) (context.Context, context.CancelFunc) {
	// Default timeout of 10 seconds for database operations
	return context.WithTimeout(parent, 10*time.Second)
}
