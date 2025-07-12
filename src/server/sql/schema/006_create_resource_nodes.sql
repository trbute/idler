-- +goose Up
CREATE TABLE resource_nodes(
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL,
	action_id INTEGER NOT NULL,
	tier INTEGER NOT NULL,
	FOREIGN KEY (action_id) REFERENCES actions (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE resource_nodes; 