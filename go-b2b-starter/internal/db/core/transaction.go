package core

import "context"

// Transaction represents a database transaction
type Transaction interface {
	Connection
	
	// Commit commits the transaction
	Commit(ctx context.Context) error
	
	// Rollback rolls back the transaction
	Rollback(ctx context.Context) error
}

// TxFunc represents a function that runs within a transaction
type TxFunc func(ctx context.Context, tx Transaction) error

// WithTransaction executes a function within a transaction
// It automatically handles commit/rollback based on the function's return value
func WithTransaction(ctx context.Context, pool Pool, fn TxFunc) error {
	tx, err := pool.BeginTx(ctx)
	if err != nil {
		return err
	}
	
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()
	
	if err := fn(ctx, tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return ErrTxRollbackFailed{
				OriginalErr: err,
				RollbackErr: rbErr,
			}
		}
		return err
	}
	
	if err := tx.Commit(ctx); err != nil {
		return ErrTxCommitFailed{Err: err}
	}
	
	return nil
}