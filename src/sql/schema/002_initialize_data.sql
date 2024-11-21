-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

INSERT INTO grid (position_x, position_y, created_at, updated_at)
VALUES (0, 0, NOW(), NOW());

INSERT INTO resource_nodes (id, name, position_x, position_y, created_at, updated_at)
VALUES (uuid_generate_v4(), 'BALSA TREE', 0, 0, NOW(), NOW());

INSERT INTO items (id, name, created_at, updated_at)
VALUES (uuid_generate_v4(), 'BALSA WOOD', Now(), Now());

INSERT INTO resources (id, resource_node_id, item_id, drop_chance, created_at, updated_at)
VALUES (
	uuid_generate_v4(), 
	(SELECT id FROM resource_nodes where name = 'BALSA TREE'), 
	(SELECT id FROM items where name = 'BALSA WOOD'),
	100,
	NOW(), 
	NOW()
);

INSERT INTO actions (id, name, created_at, updated_at) 
VALUES 
	(0, 'IDLE', NOW(), NOW()),
	(1, 'WOODCUTTING', NOW(), NOW());

-- +goose Down
DELETE FROM actions WHERE id in (0, 1);
DELETE FROM resources WHERE resource_node_id = (SELECT id FROM resource_nodes WHERE name = 'BALSA TREE') AND item_id = (SELECT id FROM items WHERE name = 'BALSA WOOD');
DELETE FROM items WHERE name = 'BALSA WOOD';
DELETE FROM resource_nodes WHERE name = 'BALSA TREE';  
DELETE FROM grid WHERE position_x = 0 AND position_y = 0;
