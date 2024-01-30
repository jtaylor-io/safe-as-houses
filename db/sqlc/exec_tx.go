package db

import (
	"context"
	"fmt"
)

// execTx executes a function within a db transaction
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.connPool.Begin(ctx)
	if err != nil {
		return err
	}

	q := New(tx)
	txErr := fn(q)
	if txErr != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("txErr: %w, rbErr: %w", txErr, rbErr)
		}
		return txErr
	}

	return tx.Commit(ctx)
}
