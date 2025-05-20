-- name: GetGrid :many
SELECT * FROM grid;

-- name: CreateGridItem :exec
INSERT INTO grid (position_x, position_y) VALUES ($1, $2);
