-- name: GetResourceNodeSpawns :many
SELECT * FROM resource_node_spawns;

-- name: GetResourceNodeSpawnsByCoordinates :many
SELECT * FROM resource_node_spawns WHERE position_x = $1 AND position_y = $2;

-- name: GetResourceNodeSpawnByCoordsAndNodeId :one
SELECT * FROM resource_node_spawns WHERE position_x = $1 AND position_y = $2 AND node_id = $3;

-- name: CreateResourceNodeSpawn :exec
INSERT INTO resource_node_spawns (node_id, position_x, position_y) VALUES ($1, $2, $3);
