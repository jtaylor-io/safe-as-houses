package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/shopspring/decimal"
)

type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error)
}

// Store provides db query and transaction capabilities
type SQLStore struct {
	Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *SQLStore {
	return &SQLStore{
		db:      db,
		Queries: *New(db),
	}
}

// execTx executes a function within a db transaction
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	q := New(tx)
	txErr := fn(q)
	if txErr != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("txErr: %w, rbErr: %w", txErr, rbErr)
		}
		return txErr
	}

	return tx.Commit()
}

// TransferTxParams contains the input parameters of the transfer transaction
type TransferTxParams struct {
	FromAccountId int64           `json:"from_account_id"`
	ToAccountId   int64           `json:"to_account_id"`
	Amount        decimal.Decimal `json:"amount"`
}

// TransferTxResult is the result of transfer tx
type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

// TransferTx performs a money transfer from one account to another
// It creates a transfer record, adds account entries and updates the accounts' balances
// within a single db tx
func (store *SQLStore) TransferTx(
	ctx context.Context,
	arg TransferTxParams,
) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountId,
			ToAccountID:   arg.ToAccountId,
			Amount:        arg.Amount,
		})
		if err != nil {
			return err
		}

		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountId,
			Amount:    arg.Amount.Neg(),
		})
		if err != nil {
			return err
		}

		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountId,
			Amount:    arg.Amount,
		})
		if err != nil {
			return err
		}

		// update accounts' balance

		if arg.FromAccountId < arg.ToAccountId {
			result.FromAccount, result.ToAccount, err = addMoney(
				ctx,
				q,
				arg.FromAccountId,
				arg.Amount.Neg(),
				arg.ToAccountId,
				arg.Amount,
			)
		} else {
			result.FromAccount, result.ToAccount, err = addMoney(ctx, q, arg.ToAccountId, arg.Amount, arg.FromAccountId, arg.Amount.Neg())
		}

		return nil
	})
	return result, err
}

func addMoney(
	ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 decimal.Decimal,
	accountID2 int64,
	amount2 decimal.Decimal,
) (Account, Account, error) {
	account1, err := q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID1,
		Amount: amount1,
	})
	if err != nil {
		return Account{}, Account{}, err
	}

	account2, err := q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})
	if err != nil {
		return Account{}, Account{}, err
	}
	return account1, account2, nil
}
