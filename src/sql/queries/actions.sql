-- name: GetActionByID :one
SELECT * FROM Actions
WHERE id = $1;

-- name: GetAllActions :many
SELECT id, name FROM ACTIONS;
