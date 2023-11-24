package db

import (
	"context"
)

// CreateUserTxParams contains the input parameters of the create user tx
type CreateUserTxParams struct {
	CreateUserParams
	AfterCreate func(user User) error
}

// CreateUserTxResult is the result of create user tx
type CreateUserTxResult struct {
	User User
}

// CreateUserTx creates a new user and executes the given function in a transaction.
func (store *SQLStore) CreateUserTx(
	ctx context.Context,
	arg CreateUserTxParams,
) (CreateUserTxResult, error) {
	var result CreateUserTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.User, err = q.CreateUser(ctx, arg.CreateUserParams)
		if err != nil {
			return err
		}

		return arg.AfterCreate(result.User)
	})

	return result, err
}
