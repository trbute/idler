-- name: GetInventoryItemsByInventoryId :many
SELECT * FROM inventory_items
WHERE inventory_id = $1;

-- name: AddItemsToInventory :one
INSERT INTO inventory_items(id, inventory_id, item_id, quantity, updated_at, created_at)
VALUES (gen_random_uuid(), $1, $2, $3, NOW(), NOW())
ON CONFLICT (inventory_id, item_id) 
DO UPDATE SET
    inventory_id = EXCLUDED.inventory_id,
    item_id = EXCLUDED.item_id,
	quantity = EXCLUDED.quantity
RETURNING *;
