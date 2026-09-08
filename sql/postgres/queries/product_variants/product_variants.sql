-- name: GetBySlug :one
SELECT * FROM product_variants WHERE slug = $1 LIMIT 1;

-- name: GetByProductID :many
SELECT * FROM product_variants WHERE product_id = $1 ORDER BY sort_order ASC;

-- name: SitemapProducts :many
-- Enabled product-variant pages for the sitemap feed. Both the variant and its
-- parent product must be enabled; updated_at is the variant's own last change.
SELECT pv.slug, pv.updated_at
FROM product_variants pv
JOIN products p ON p.id = pv.product_id
WHERE pv.is_enable AND p.is_enable
ORDER BY pv.updated_at DESC NULLS LAST;
