-- name: CreateTransfer :one
INSERT INTO transfers (
    from_account_id,
    to_account_id,
    amount
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetTransfer :one
SELECT * FROM transfers
WHERE id = $1 LIMIT 1;

-- name: GetTransferForUpdate :one
SELECT * FROM transfers
WHERE id = $1 LIMIT 1
FOR NO KEY UPDATE;

-- name: DeleteTransfer :exec
DELETE FROM transfers
WHERE id = $1;
