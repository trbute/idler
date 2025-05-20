-- +goose Up
CREATE TABLE resources(
	id SERIAL PRIMARY KEY,
	resource_node_id INTEGER NOT NULL,
	item_id INTEGER NOT NULL,
	drop_chance INTEGER NOT NULL,
	FOREIGN KEY (item_id) REFERENCES items (id) ON DELETE CASCADE,
	FOREIGN KEY (resource_node_id) REFERENCES resource_nodes (id) ON DELETE CASCADE 
);

-- +goose Down
DROP TABLE resources; 