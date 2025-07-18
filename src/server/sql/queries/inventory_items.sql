-- name: GetInventoryItemsByInventoryId :many
SELECT * FROM inventory_items
WHERE inventory_id = $1;

-- name: AddItemsToInventory :one
INSERT INTO inventory_items(id, inventory_id, item_id, quantity, updated_at, created_at)
VALUES (gen_random_uuid(), $1, $2, $3, NOW(), NOW())
ON CONFLICT (inventory_id, item_id) 
DO UPDATE SET
	quantity = inventory_items.quantity + EXCLUDED.quantity,
	updated_at = NOW()
RETURNING *;

-- name: BatchAddItemsToInventory :exec
INSERT INTO inventory_items(id, inventory_id, item_id, quantity, updated_at, created_at)
SELECT gen_random_uuid(), unnest($1::UUID[]), unnest($2::INTEGER[]), unnest($3::INTEGER[]), NOW(), NOW()
ON CONFLICT (inventory_id, item_id) 
DO UPDATE SET
	quantity = inventory_items.quantity + EXCLUDED.quantity,
	updated_at = NOW();

-- name: RemoveItemsFromInventory :exec
UPDATE inventory_items 
SET quantity = quantity - $3, updated_at = NOW()
WHERE inventory_id = $1 AND item_id = $2 AND quantity >= $3;

-- name: GetInventoryItemQuantity :one
SELECT quantity FROM inventory_items
WHERE inventory_id = $1 AND item_id = $2;

-- name: DeleteEmptyInventoryItems :exec
DELETE FROM inventory_items
WHERE inventory_id = $1 AND quantity <= 0;
