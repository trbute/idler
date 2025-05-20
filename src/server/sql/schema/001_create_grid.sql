-- +goose Up
CREATE TABLE grid(
	position_x INTEGER NOT NULL,
	position_y INTEGER NOT NULL,
	PRIMARY KEY (position_x, position_y) 
);

-- +goose Down
DROP TABLE grid; 