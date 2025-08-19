-- name: DeleteByAttributeGroupID :exec
DELETE FROM attributes WHERE attribute_group_id = $1::uuid;

-- name: GetByID :one
SELECT * FROM attributes WHERE id = $1 LIMIT 1;