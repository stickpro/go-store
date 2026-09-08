-- name: GetByID :one
SELECT * FROM categories WHERE id = $1 LIMIT 1;

-- name: GetBySlug :one
SELECT * FROM categories WHERE slug = $1 LIMIT 1;

-- name: GetAllForTree :many
SELECT * FROM categories
WHERE is_enable = true
ORDER BY id ASC ;


-- name: SitemapCategories :many
-- Enabled category pages for the sitemap feed.
SELECT slug, updated_at
FROM categories
WHERE is_enable = true
ORDER BY updated_at DESC NULLS LAST;
