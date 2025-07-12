-- +goose Up
CREATE TABLE actions(
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL,
	required_tool_id INTEGER REFERENCES items(id)
);

-- +goose Down
DROP TABLE actions; 