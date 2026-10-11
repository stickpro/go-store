-- name: GetByCity :many
-- Settlements inside a city carry the parent city name too, so only rows that
-- are a city (or a federal-city region) match; the federal city itself sorts
-- first by fias_level when population is unknown.
SELECT * FROM cities
WHERE (city=sqlc.arg(name)::text OR region=sqlc.arg(name)::text) AND settlement IS NULL
ORDER BY population DESC, fias_level;

-- name: GetCityOrderByPopulation :many
SELECT * FROM cities ORDER BY population DESC LIMIT 20;
