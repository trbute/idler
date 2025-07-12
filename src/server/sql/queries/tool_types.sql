-- name: GetToolTypeById :one
SELECT * FROM tool_types WHERE id = $1;