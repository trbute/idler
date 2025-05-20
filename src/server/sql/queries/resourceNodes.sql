-- name: GetResourceNodeById :one
SELECT * FROM resource_nodes WHERE id = $1;

-- name: CreateResourceNode :exec
INSERT INTO resource_nodes (name, action_id, tier) VALUES ($1, $2, $3);

-- name: GetResourceNodeByName :one
SELECT * FROM resource_nodes WHERE name = $1;
