-- name: GetActionByID :one
SELECT * FROM actions
WHERE id = $1;

-- name: GetAllActions :many
SELECT id, name FROM actions;
