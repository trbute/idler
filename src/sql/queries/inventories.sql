-- name: CreateInventory :one
INSERT INTO inventories(id, character_id, position_x, position_y, created_at, updated_at)
VALUES (
	gen_random_uuid(),
	$1,
	$2,
	$3,
	NOW(),
	NOW()
)
RETURNING *;

-- name: GetInventory :one
SELECT * FROM inventories
WHERE id = $1;

-- name: GetInventoryByUserId :one
SELECT * FROM inventories
WHERE character_id = $1;
