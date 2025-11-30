-- name: GetBreadcrumbsByProductSlug :many
SELECT
    c.id,
    c.name,
    c.slug,
    c.meta_title,
    c.meta_h1,
    cp.depth
FROM products p
JOIN categories c_product ON p.category_id = c_product.id
JOIN category_paths cp ON cp.descendant_id = c_product.id
JOIN categories c ON c.id = cp.ancestor_id
WHERE p.slug = $1
ORDER BY cp.depth ASC;

-- name: GetBreadcrumbsByCategoryID :many
SELECT
    c.id,
    c.name,
    c.slug,
    c.meta_title,
    c.meta_h1,
    cp.depth
FROM category_paths cp
JOIN categories c ON c.id = cp.ancestor_id
WHERE cp.descendant_id = $1
ORDER BY cp.depth ASC;

-- name: GetBreadcrumbsByCategorySlug :many
SELECT
    c.id,
    c.name,
    c.slug,
    c.meta_title,
    c.meta_h1,
    cp.depth
FROM categories target_cat
JOIN category_paths cp ON cp.descendant_id = target_cat.id
JOIN categories c ON c.id = cp.ancestor_id
WHERE target_cat.slug = $1
ORDER BY cp.depth ASC;

-- name: GetAllDescendants :many
SELECT
    c.id,
    c.parent_id,
    c.name,
    c.slug,
    cp.depth
FROM category_paths cp
JOIN categories c ON c.id = cp.descendant_id
WHERE cp.ancestor_id = $1
ORDER BY cp.depth ASC, c.name ASC;

-- name: GetDirectChildren :many
SELECT
    c.id,
    c.parent_id,
    c.name,
    c.slug,
    c.is_enable
FROM category_paths cp
JOIN categories c ON c.id = cp.descendant_id
WHERE cp.ancestor_id = $1
  AND cp.depth = 1
ORDER BY c.name ASC;

-- name: GetAllAncestors :many
SELECT
    c.id,
    c.name,
    c.slug,
    cp.depth
FROM category_paths cp
JOIN categories c ON c.id = cp.ancestor_id
WHERE cp.descendant_id = $1
ORDER BY cp.depth DESC;

-- name: GetCategoryDepth :one
SELECT COALESCE(MAX(depth), 0) as depth
FROM category_paths
WHERE descendant_id = $1;

-- name: IsCategoryDescendantOf :one
SELECT EXISTS(
    SELECT 1
    FROM category_paths
    WHERE ancestor_id = $1
      AND descendant_id = $2
      AND depth > 0
) as is_descendant;

-- name: DeleteCategoryPaths :exec
DELETE FROM category_paths
WHERE descendant_id IN (
    SELECT cp1.descendant_id
    FROM category_paths cp1
    WHERE cp1.ancestor_id = $1
)
AND ancestor_id IN (
    SELECT cp2.ancestor_id
    FROM category_paths cp2
    WHERE cp2.descendant_id = $1
    AND cp2.depth > 0
);


-- name: InsertCategoryPath :exec
INSERT INTO category_paths (ancestor_id, descendant_id, depth)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: GetCategoryPathsBatch :many
SELECT
    cp.ancestor_id,
    cp.descendant_id,
    cp.depth,
    c.name as descendant_name,
    c.slug as descendant_slug
FROM category_paths cp
JOIN categories c ON c.id = cp.descendant_id
WHERE cp.descendant_id = ANY($1::uuid[])
ORDER BY cp.descendant_id, cp.depth ASC;
