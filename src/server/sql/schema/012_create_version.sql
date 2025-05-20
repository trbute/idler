-- +goose Up
CREATE TABLE version(
	id SERIAL PRIMARY KEY,
	value TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL
);

INSERT INTO version (value, created_at, updated_at) VALUES ('0.0.0', NOW(), NOW());

-- +goose Down
DROP TABLE version; 