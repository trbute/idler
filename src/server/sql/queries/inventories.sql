-- name: CreateInventory :one
INSERT INTO inventories(id, character_id, position_x, position_y, capacity, created_at, updated_at)
VALUES (
	gen_random_uuid(),
	$1,
	$2,
	$3,
	$4,
	NOW(),
	NOW()
)
RETURNING *;

-- name: GetInventory :one
SELECT * FROM inventories
WHERE id = $1;

-- name: GetInventoryByCharacterId :one
SELECT * FROM inventories
WHERE character_id = $1;

-- name: UpdateInventoryWeight :exec
UPDATE inventories
SET weight = weight + $2, updated_at = NOW()
WHERE id = $1;
