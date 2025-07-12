-- +goose Up
ALTER TABLE users RENAME COLUMN username TO surname;

-- +goose Down
ALTER TABLE users RENAME COLUMN surname TO username;