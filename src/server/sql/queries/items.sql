-- name: GetItemByResourceId :one
SELECT * FROM items
WHERE id = (SELECT item_id FROM resources WHERE resources.id = $1);

-- name: GetItemById :one
SELECT * FROM items
WHERE id = $1;
