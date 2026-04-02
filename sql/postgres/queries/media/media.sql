-- name: Get :one
SELECT * FROM media WHERE id = $1 LIMIT 1;

-- name: GetBySourceURL :one
SELECT * FROM media WHERE source_url = $1 LIMIT 1;

-- name: CreateWithSourceURL :one
INSERT INTO media (name, path, file_name, mime_type, disk_type, size, source_url, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, now())
RETURNING *;