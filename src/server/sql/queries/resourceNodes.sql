-- name: GetResourceNodeById :one
SELECT * FROM resource_nodes WHERE id = $1;
