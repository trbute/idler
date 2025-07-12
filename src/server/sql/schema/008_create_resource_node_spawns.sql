-- +goose Up
CREATE TABLE resource_node_spawns(
	id SERIAL PRIMARY KEY,
	node_id INTEGER NOT NULL,
	position_x INTEGER NOT NULL,
	position_y INTEGER NOT NULL,
	FOREIGN KEY (position_x, position_y) REFERENCES grid (position_x, position_y) ON DELETE CASCADE, 
	FOREIGN KEY (node_id) REFERENCES resource_nodes (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE resource_node_spawns; 