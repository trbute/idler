// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: tool_types.sql

package database

import (
	"context"
)

const getToolTypeById = `-- name: GetToolTypeById :one
SELECT id, name, tier FROM tool_types WHERE id = $1
`

func (q *Queries) GetToolTypeById(ctx context.Context, id int32) (ToolType, error) {
	row := q.db.QueryRow(ctx, getToolTypeById, id)
	var i ToolType
	err := row.Scan(&i.ID, &i.Name, &i.Tier)
	return i, err
}
