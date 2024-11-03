-- name: CreateInventory :one
INSERT INTO inventories(id, user_id, created_at, updated_at)
VALUES (
	gen_random_uuid(),
	$1,
	NOW(),
	NOW()
)
RETURNING *;

-- name: GetInventory :one
SELECT * from inventories
WHERE id = $1;
