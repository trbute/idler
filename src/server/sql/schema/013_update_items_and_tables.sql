-- +goose Up
ALTER TABLE items ADD COLUMN tool_type_id INTEGER REFERENCES tool_types(id);
ALTER TABLE actions DROP COLUMN required_tool_id;
ALTER TABLE actions ADD COLUMN required_tool_type_id INTEGER REFERENCES tool_types(id);
ALTER TABLE resource_nodes ADD COLUMN min_tool_tier INTEGER NOT NULL DEFAULT 1;

-- +goose Down
ALTER TABLE resource_nodes DROP COLUMN min_tool_tier;
ALTER TABLE actions DROP COLUMN required_tool_type_id;
ALTER TABLE actions ADD COLUMN required_tool_id INTEGER REFERENCES items(id);
ALTER TABLE items DROP COLUMN tool_type_id;