-- name: Create :one
INSERT INTO users (email, email_verified_at, password, remember_token, location, language, created_at, deleted_at, banned)
	VALUES ($1, $2, $3, $4, $5, $6, now(), $7, $8)
	RETURNING *;

-- name: Delete :exec
DELETE FROM users WHERE id=$1;

-- name: GetAll :many
SELECT * FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: Update :one
UPDATE users
	SET location=$1, language=$2, updated_at=$3, banned=$4
	WHERE id=$5
	RETURNING *;

