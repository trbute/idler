-- +goose Up
CREATE TABLE actions(
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL
);

-- +goose Down
DROP TABLE actions; 