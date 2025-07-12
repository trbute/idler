-- +goose Up
ALTER TABLE users ADD COLUMN username VARCHAR(20) UNIQUE;

-- +goose Down
ALTER TABLE users DROP COLUMN username;