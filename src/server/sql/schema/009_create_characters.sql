-- +goose Up
CREATE TABLE characters(
	id UUID PRIMARY KEY,
	user_id UUID NOT NULL REFERENCES users ON DELETE CASCADE,
	name VARCHAR(20) UNIQUE NOT NULL,
	position_x INTEGER NOT NULL DEFAULT 0,
	position_y INTEGER NOT NULL DEFAULT 0,
	action_id INTEGER NOT NULL REFERENCES actions ON DELETE CASCADE DEFAULT 1,
	action_target INTEGER REFERENCES resource_node_spawns ON DELETE CASCADE DEFAULT NULL,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	FOREIGN KEY (position_x, position_y) REFERENCES grid (position_x, position_y) ON DELETE CASCADE 
);

-- +goose Down
DROP TABLE characters; 