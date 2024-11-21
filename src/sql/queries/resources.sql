-- name: GetResourcesByNodeId :many
SELECT * FROM resources
WHERE resource_node_id = $1;
