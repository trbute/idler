-- +goose Up
CREATE TABLE tool_types(
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	tier INTEGER NOT NULL
);

-- +goose Down
DROP TABLE tool_types;