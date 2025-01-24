-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

INSERT INTO grid (position_x, position_y)
VALUES 
	(0, 0)
;

INSERT INTO actions (name) 
VALUES 
	('IDLE'),
	('MELEE'),
	('ARCHERY'),
	('WEAVING'),
	('GATHERING'),
	('WOODCUTTING'),
	('STONEBREAKING'),
	('MINING')
;

INSERT INTO resource_nodes (name, action_id, tier)
VALUES 
	('STICKS', 4, 1),
	('ROCKS', 4, 1),
	('BALSA TREE', 5, 1),
	('SOAPSTONE DEPOSIT', 6, 1)
;

INSERT INTO items (name, weight)
VALUES 
	('STICKS', 1),
	('ROCKS', 1),
	('BALSA LOGS', 2),
	('SOAPSTONE', 1)
;

INSERT INTO resources (resource_node_id, item_id, drop_chance)
VALUES 
	(
		(SELECT id FROM resource_nodes where name = 'STICKS'), 
		(SELECT id FROM items where name = 'STICKS'),
		100
	),
	(
		(SELECT id FROM resource_nodes where name = 'ROCKS'), 
		(SELECT id FROM items where name = 'ROCKS'),
		100
	),
	(
		(SELECT id FROM resource_nodes where name = 'BALSA TREE'), 
		(SELECT id FROM items where name = 'BALSA LOGS'),
		100
	),
	(
		(SELECT id FROM resource_nodes where name = 'SOAPSTONE DEPOSIT'), 
		(SELECT id FROM items where name = 'SOAPSTONE'),
		100
	)
;

INSERT INTO resource_node_spawns (node_id, position_x, position_y)
VALUES 
	(
		(SELECT id FROM resource_nodes where name = 'STICKS'), 
		0,
		0
	),
	(
		(SELECT id FROM resource_nodes where name = 'ROCKS'), 
		0,
		0
	),
	(
		(SELECT id FROM resource_nodes where name = 'BALSA TREE'), 
		0,
		0
	),
	(
		(SELECT id FROM resource_nodes where name = 'SOAPSTONE DEPOSIT'), 
		0,
		0
	)
;


-- +goose Down
DELETE FROM actions;
DELETE FROM resources;
DELETE FROM items;
DELETE FROM resource_nodes;  
DELETE FROM resource_node_spawns;  
DELETE FROM grid;
