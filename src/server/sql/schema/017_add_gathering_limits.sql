-- +goose Up
ALTER TABLE characters ADD COLUMN action_amount_limit INTEGER DEFAULT NULL;
ALTER TABLE characters ADD COLUMN action_amount_progress INTEGER DEFAULT 0;

-- +goose Down
ALTER TABLE characters DROP COLUMN action_amount_limit;
ALTER TABLE characters DROP COLUMN action_amount_progress;