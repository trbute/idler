-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

INSERT INTO grid (position_x, position_y, created_at, updated_at)
VALUES (0, 0, NOW(), NOW());

INSERT INTO resources (id, name, created_at, updated_at)
VALUES (uuid_generate_v4(), 'BALSA TREE', NOW(), NOW());

INSERT INTO resource_nodes (id, resource_id, position_x, position_y, created_at, updated_at)
VALUES (uuid_generate_v4(), (SELECT id from resources WHERE name = 'BALSA TREE'), 0, 0, NOW(), NOW());

INSERT INTO items (id, name, created_at, updated_at)
VALUES (uuid_generate_v4(), 'BALSA WOOD', Now(), Now());

INSERT INTO resource_drops (id, resource_id, item_id, drop_chance, created_at, updated_at)
VALUES (uuid_generate_v4(), (SELECT id from resources WHERE name = 'BALSA TREE'), (SELECT id from items WHERE name = 'BALSA WOOD'), 100.0, NOW(), NOW());

INSERT INTO actions (id, name, created_at, updated_at) 
VALUES 
	(0, 'IDLE', NOW(), NOW()),
	(1, 'WOODCUTTING', NOW(), NOW());

-- +goose Down
DELETE FROM grid WHERE position_x = 0 AND position_y = 0;
DELETE FROM resources where name = 'BALSA TREE';
DELETE FROM resource_nodes WHERE resource_id = (SELECT id FROM resources where name = 'BALSA TREE');  
DELETE FROM items WHERE name = 'BALSA WOOD';
DELETE FROM resource_drops WHERE resource_id = (SELECT id FROM resources where name = 'BALSA TREE');
DELETE FROM actions WHERE id in (0, 1);
