-- +goose Up
CREATE TABLE grid(
	position_x INTEGER NOT NULL,
	position_y INTEGER NOT NULL,
	PRIMARY KEY (position_x, position_y) 
);

CREATE TABLE users(
	id UUID PRIMARY KEY,
	email TEXT UNIQUE NOT NULL, 
	hashed_password TEXT NOT NULL DEFAULT 'unset',
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL
);

CREATE TABLE refresh_tokens(
	token TEXT PRIMARY KEY,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	user_id UUID NOT NULL REFERENCES users ON DELETE CASCADE,
	expires_at TIMESTAMP NOT NULL,
	revoked_at TIMESTAMP DEFAULT NULL
);

CREATE TABLE actions(
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL
);

CREATE TABLE characters(
	id UUID PRIMARY KEY,
	user_id UUID NOT NULL REFERENCES users ON DELETE CASCADE,
	name VARCHAR(20) UNIQUE NOT NULL,
	position_x INTEGER NOT NULL DEFAULT 0,
	position_y INTEGER NOT NULL DEFAULT 0,
	action_id INTEGER NOT NULL REFERENCES actions ON DELETE CASCADE DEFAULT 0,
	action_target UUID,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	FOREIGN KEY (position_x, position_y) REFERENCES grid (position_x, position_y) ON DELETE CASCADE 
);

CREATE TABLE items(
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL
);

CREATE TABLE inventories(
	id UUID PRIMARY KEY,
	character_id UUID,
	position_x INTEGER NOT NULL,
	position_y INTEGER NOT NULL,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	FOREIGN KEY (position_x, position_y) REFERENCES grid (position_x, position_y) ON DELETE CASCADE,
	FOREIGN KEY (character_id) REFERENCES characters (id) ON DELETE CASCADE 
);

CREATE TABLE inventory_items(
	id UUID PRIMARY KEY,
	item_id INTEGER NOT NULL,
	inventory_id UUID NOT NULL,
	quantity INTEGER NOT NULL,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	FOREIGN KEY (item_id) REFERENCES items (id) ON DELETE CASCADE,
	FOREIGN KEY (inventory_id) REFERENCES inventories (id) ON DELETE CASCADE,
	UNIQUE(item_id, inventory_id)
);

CREATE TABLE resource_nodes(
	id INTEGER PRIMARY KEY,
	name TEXT,
	action_id INTEGER NOT NULL,
	tier INTEGER NOT NULL,
	FOREIGN KEY (action_id) REFERENCES actions (id) ON DELETE CASCADE
);

CREATE TABLE resources(
	id INTEGER PRIMARY KEY,
	resource_node_id INTEGER NOT NULL,
	item_id INTEGER NOT NULL,
	drop_chance INTEGER NOT NULL,
	FOREIGN KEY (item_id) REFERENCES items (id) ON DELETE CASCADE,
	FOREIGN KEY (resource_node_id) REFERENCES resource_nodes (id) ON DELETE CASCADE 

);

CREATE TABLE resource_node_spawns(
	id INTEGER PRIMARY KEY,
	node_id INTEGER NOT NULL,
	position_x INTEGER NOT NULL,
	position_y INTEGER NOT NULL,
	FOREIGN KEY (position_x, position_y) REFERENCES grid (position_x, position_y) ON DELETE CASCADE, 
	FOREIGN KEY (node_id) REFERENCES resource_nodes (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE resources;
DROP TABLE resource_nodes;
DROP TABLE inventory_items;
DROP TABLE inventories;
DROP TABLE items;
DROP TABLE characters;
DROP TABLE actions;
DROP TABLE refresh_tokens;
DROP TABLE users;
DROP TABLE grid;
