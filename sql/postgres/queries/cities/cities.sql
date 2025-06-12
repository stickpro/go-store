-- name: GetByCity :many
SELECT * FROM cities WHERE city=$1 or region=$1;

-- name: GetCityOrderByPopulation :many
SELECT * FROM cities ORDER BY population DESC LIMIT 20;
