package core

import (
	"context"
	"time"
)

// Connection represents a database connection interface
type Connection interface {
	// Execute runs a query that doesn't return rows
	Execute(ctx context.Context, query string, args ...any) error

	// Query executes a query that returns rows
	Query(ctx context.Context, query string, args ...any) (Rows, error)

	// QueryRow executes a query that returns a single row
	QueryRow(ctx context.Context, query string, args ...any) Row

	// BeginTx starts a new transaction
	BeginTx(ctx context.Context) (Transaction, error)

	// Ping verifies the connection to the database is still alive
	Ping(ctx context.Context) error

	// Close closes the database connection
	Close() error
}

// Pool represents a connection pool interface
type Pool interface {
	Connection

	// Stats returns connection pool statistics
	Stats() PoolStats

	// SetMaxConnections sets the maximum number of connections in the pool
	SetMaxConnections(n int)

	// SetMaxConnectionLifetime sets the maximum lifetime of a connection
	SetMaxConnectionLifetime(d time.Duration)

	// SetMaxConnectionIdleTime sets the maximum idle time of a connection
	SetMaxConnectionIdleTime(d time.Duration)
}

// PoolStats contains connection pool statistics
type PoolStats struct {
	TotalConnections    int
	IdleConnections     int
	AcquiredConnections int
	MaxConnections      int
}

// Rows represents the result of a query
type Rows interface {
	// Next prepares the next row for reading
	Next() bool

	// Scan reads the values from the current row
	Scan(dest ...any) error

	// Close closes the rows
	Close() error

	// Err returns any error that occurred during iteration
	Err() error
}

// Row represents a single row result
type Row interface {
	// Scan reads the values from the row
	Scan(dest ...any) error
}
