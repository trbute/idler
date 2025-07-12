-- +goose Up
CREATE TABLE items(
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL,
	weight INTEGER NOT NULL
);

-- +goose Down
DROP TABLE items; 