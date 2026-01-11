-- name: DeleteByAttributeGroupID :exec
DELETE FROM attributes WHERE attribute_group_id = $1::uuid;

-- name: GetByID :one
SELECT * FROM attributes WHERE id = $1 LIMIT 1;

-- name: GetBySlug :one
SELECT * FROM attributes WHERE slug = $1 LIMIT 1;

-- name: GetFilterableAttributes :many
SELECT
    a.id,
    a.attribute_group_id,
    a.name,
    a.slug,
    a.type,
    a.unit,
    a.is_filterable,
    a.is_visible,
    a.is_required,
    a.sort_order,
    a.created_at,
    a.updated_at,
    ag.name as group_name,
    ag.slug as group_slug
FROM attributes a
INNER JOIN attribute_groups ag ON a.attribute_group_id = ag.id
WHERE a.is_filterable = true
ORDER BY ag.name ASC, a.sort_order ASC, a.name ASC;
