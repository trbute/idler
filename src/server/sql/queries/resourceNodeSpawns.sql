-- name: GetResourceNodeSpawns :many
SELECT * FROM resource_node_spawns;

-- name: CreateResourceNodeSpawn :exec
INSERT INTO resource_node_spawns (node_id, position_x, position_y) VALUES ($1, $2, $3);
