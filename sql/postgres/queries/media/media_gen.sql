-- name: Create :one
INSERT INTO media (name, path, file_name, mime_type, disk_type, size, created_at, source_url)
	VALUES ($1, $2, $3, $4, $5, $6, now(), $7)
	RETURNING *;

-- name: Delete :exec
DELETE FROM media WHERE id=$1;

