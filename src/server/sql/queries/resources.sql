-- name: GetResourcesByNodeId :many
SELECT * FROM resources
WHERE resource_node_id = $1;

-- name: CreateResource :exec
INSERT INTO resources (resource_node_id, item_id, drop_chance) VALUES ($1, $2, $3);
