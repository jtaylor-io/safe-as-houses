package db

import (
	"context"
	"database/sql"
)

// VerifyEmailTxParams contains the input parameters of the verify email tx
type VerifyEmailTxParams struct {
	EmailId    int64
	SecretCode string
}

// VerifyEmailTxResult is the result of verify email tx
type VerifyEmailTxResult struct {
	User        User
	VerifyEmail VerifyEmail
}

// VerifyEmailTx verify email for user
func (store *SQLStore) VerifyEmailTx(
	ctx context.Context,
	arg VerifyEmailTxParams,
) (VerifyEmailTxResult, error) {
	var result VerifyEmailTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error
		result.VerifyEmail, err = q.UpdateVerifyEmail(ctx, UpdateVerifyEmailParams{
			ID:         arg.EmailId,
			SecretCode: arg.SecretCode,
		})
		if err != nil {
			return err
		}

		result.User, err = q.UpdateUser(ctx, UpdateUserParams{
			IsEmailVerified: sql.NullBool{
				Bool:  true,
				Valid: true,
			},
			Username: result.VerifyEmail.Username,
		})
		return err
	})

	return result, err
}
