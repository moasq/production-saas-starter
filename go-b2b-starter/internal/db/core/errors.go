package core

import (
	"errors"
	"fmt"
)

// Common database errors
var (
	// ErrNoRows is returned when a query returns no rows
	ErrNoRows = errors.New("no rows in result set")
	
	// ErrTxClosed is returned when an operation is attempted on a closed transaction
	ErrTxClosed = errors.New("transaction has already been committed or rolled back")
	
	// ErrPoolClosed is returned when an operation is attempted on a closed pool
	ErrPoolClosed = errors.New("connection pool is closed")
	
	// ErrInvalidConnection is returned when the connection is invalid
	ErrInvalidConnection = errors.New("invalid database connection")
	
	// ErrTimeout is returned when a database operation times out
	ErrTimeout = errors.New("database operation timed out")
)

// ErrTxRollbackFailed is returned when a transaction rollback fails
type ErrTxRollbackFailed struct {
	OriginalErr error
	RollbackErr error
}

func (e ErrTxRollbackFailed) Error() string {
	return fmt.Sprintf("transaction rollback failed: %v (original error: %v)", e.RollbackErr, e.OriginalErr)
}

func (e ErrTxRollbackFailed) Unwrap() error {
	return e.OriginalErr
}

// ErrTxCommitFailed is returned when a transaction commit fails
type ErrTxCommitFailed struct {
	Err error
}

func (e ErrTxCommitFailed) Error() string {
	return fmt.Sprintf("transaction commit failed: %v", e.Err)
}

func (e ErrTxCommitFailed) Unwrap() error {
	return e.Err
}

// ErrConstraintViolation represents a database constraint violation
type ErrConstraintViolation struct {
	Constraint string
	Message    string
}

func (e ErrConstraintViolation) Error() string {
	return fmt.Sprintf("constraint violation '%s': %s", e.Constraint, e.Message)
}

// IsNoRowsError checks if an error is a no rows error
func IsNoRowsError(err error) bool {
	return errors.Is(err, ErrNoRows)
}

// IsConstraintError checks if an error is a constraint violation
func IsConstraintError(err error) bool {
	var constraintErr ErrConstraintViolation
	return errors.As(err, &constraintErr)
}

// IsTimeoutError checks if an error is a timeout error
func IsTimeoutError(err error) bool {
	return errors.Is(err, ErrTimeout)
}