// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: resourceNodes.sql

package database

import (
	"context"
)

const getResourceNodes = `-- name: GetResourceNodes :many
SELECT id, name, position_x, position_y, created_at, updated_at FROM resource_nodes
`

func (q *Queries) GetResourceNodes(ctx context.Context) ([]ResourceNode, error) {
	rows, err := q.db.QueryContext(ctx, getResourceNodes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ResourceNode
	for rows.Next() {
		var i ResourceNode
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.PositionX,
			&i.PositionY,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}