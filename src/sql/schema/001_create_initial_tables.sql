-- +goose Up
CREATE TABLE grid(
	position_x INTEGER NOT NULL,
	position_y INTEGER NOT NULL,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
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
	id UUID PRIMARY KEY,
	name TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL
);

CREATE TABLE spawned_items(
	id UUID PRIMARY KEY,
	inventory UUID NOT NULL,
	position_x INTEGER NOT NULL,
	position_y INTEGER NOT NULL,
	quantity INTEGER NOT NULL,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	FOREIGN KEY (position_x, position_y) REFERENCES grid (position_x, position_y) ON DELETE CASCADE 
);

CREATE TABLE inventories(
	id UUID PRIMARY KEY,
	user_id UUID,
	position_x INTEGER NOT NULL,
	position_y INTEGER NOT NULL,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	FOREIGN KEY (position_x, position_y) REFERENCES grid (position_x, position_y) ON DELETE CASCADE 
);

CREATE TABLE resources(
	id UUID PRIMARY KEY,
	name TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL
);

CREATE TABLE resource_nodes(
	id UUID PRIMARY KEY,
	resource_id UUID NOT NULL,
	position_x INTEGER NOT NULL,
	position_y INTEGER NOT NULL,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,	
	FOREIGN KEY (position_x, position_y) REFERENCES grid (position_x, position_y) ON DELETE CASCADE 
);

CREATE TABLE resource_drops(
	id UUID PRIMARY KEY,
	resource_id UUID NOT NULL,
	item_id UUID NOT NULL,
	drop_chance DECIMAL NOT NULL,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE resource_drops;
DROP TABLE resource_nodes;
DROP TABLE resources;
DROP TABLE inventories;
DROP TABLE spawned_items;
DROP TABLE items;
DROP TABLE characters;
DROP TABLE actions;
DROP TABLE refresh_tokens;
DROP TABLE users;
DROP TABLE grid;
