-- name: Create :one
INSERT INTO attribute_values (
    attribute_id,
    value,
    value_normalized,
    value_numeric,
    display_order,
    is_active,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, now()
) RETURNING *;

-- name: GetByID :one
SELECT * FROM attribute_values
WHERE id = $1 LIMIT 1;

-- name: GetByAttributeID :many
SELECT * FROM attribute_values
WHERE attribute_id = $1 AND is_active = true
ORDER BY display_order ASC, value ASC;

-- name: GetOrCreate :one
INSERT INTO attribute_values (
    attribute_id,
    value,
    value_normalized,
    value_numeric,
    display_order,
    is_active,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, now()
)
ON CONFLICT (attribute_id, value)
DO UPDATE SET
    value_normalized = EXCLUDED.value_normalized,
    value_numeric = EXCLUDED.value_numeric,
    display_order = EXCLUDED.display_order,
    is_active = EXCLUDED.is_active,
    updated_at = now()
RETURNING *;

-- name: Update :one
UPDATE attribute_values
SET
    value = $1,
    value_normalized = $2,
    value_numeric = $3,
    display_order = $4,
    is_active = $5,
    updated_at = now()
WHERE id = $6
RETURNING *;

-- name: Delete :exec
DELETE FROM attribute_values
WHERE id = $1;

-- name: GetAll :many
SELECT * FROM attribute_values
WHERE attribute_id = $1
ORDER BY display_order ASC, value ASC
LIMIT $2 OFFSET $3;

-- name: GetWithUsageCount :many
SELECT
    av.id,
    av.attribute_id,
    av.value,
    av.value_normalized,
    av.value_numeric,
    av.display_order,
    av.is_active,
    av.created_at,
    av.updated_at,
    COUNT(DISTINCT pav.product_id) as usage_count
FROM attribute_values av
LEFT JOIN product_attribute_values pav ON av.id = pav.attribute_value_id
WHERE av.attribute_id = $1
GROUP BY av.id, av.attribute_id, av.value, av.value_normalized, av.value_numeric, av.display_order, av.is_active, av.created_at, av.updated_at
ORDER BY av.display_order ASC, av.value ASC;
