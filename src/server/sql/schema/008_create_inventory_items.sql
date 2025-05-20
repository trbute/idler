-- +goose Up
CREATE TABLE inventory_items(
	id UUID PRIMARY KEY,
	item_id INTEGER NOT NULL,
	inventory_id UUID NOT NULL,
	quantity INTEGER NOT NULL,
	created_at TIMESTAMP NOT NULL,
	updated_at TIMESTAMP NOT NULL,
	FOREIGN KEY (item_id) REFERENCES items (id) ON DELETE CASCADE,
	FOREIGN KEY (inventory_id) REFERENCES inventories (id) ON DELETE CASCADE,
	UNIQUE(item_id, inventory_id)
);

-- +goose Down
DROP TABLE inventory_items; 