-- +goose Up
CREATE TABLE inventories(
	id UUID PRIMARY KEY,
	character_id UUID,
	position_x INTEGER NOT NULL,
	position_y INTEGER NOT NULL,
	weight INTEGER NOT NULL DEFAULT 0,
	capacity INTEGER NOT NULL,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	FOREIGN KEY (position_x, position_y) REFERENCES grid (position_x, position_y) ON DELETE CASCADE,
	FOREIGN KEY (character_id) REFERENCES characters (id) ON DELETE CASCADE 
);

-- +goose Down
DROP TABLE inventories; 