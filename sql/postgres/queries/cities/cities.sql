-- name: GetByCity :many
SELECT * FROM cities WHERE city=$1 or region=$1;


