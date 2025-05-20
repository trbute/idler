-- name: GetVersion :one
SELECT * FROM version WHERE id = 1;

-- name: CreateVersion :exec
INSERT INTO version (value, created_at, updated_at) VALUES ($1, NOW(), NOW());

-- name: UpdateVersion :exec
UPDATE version SET value = $1 WHERE id = 1;