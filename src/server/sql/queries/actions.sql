-- name: GetActionByName :one
SELECT * FROM actions
WHERE name = $1;

-- name: GetAllActions :many
SELECT id, name FROM actions;
