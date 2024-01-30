-- name: CreateVerifyEmail :one
INSERT INTO verify_emails(
  username,
  email,
  secret_code
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: UpdateVerifyEmail :one
UPDATE verify_emails 
SET
  is_used = true
WHERE 
  id = @id
  AND secret_code = @secret_code
  AND is_used = FALSE
  AND expired_at > now()
RETURNING *;
