-- name: Create :one
INSERT INTO categories (parent_id, name, slug, description, meta_title, meta_h1, meta_description, meta_keyword, is_enable, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now())
	RETURNING *;

