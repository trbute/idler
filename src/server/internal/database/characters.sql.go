// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: characters.sql

package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createCharacter = `-- name: CreateCharacter :one
INSERT INTO characters(id, user_id, name, created_at, updated_at)
VALUES (
	gen_random_uuid(),
	$1,
	$2,
	NOW(),
	NOW()
)
RETURNING id, user_id, name, position_x, position_y, action_id, action_target, created_at, updated_at
`

type CreateCharacterParams struct {
	UserID pgtype.UUID
	Name   string
}

func (q *Queries) CreateCharacter(ctx context.Context, arg CreateCharacterParams) (Character, error) {
	row := q.db.QueryRow(ctx, createCharacter, arg.UserID, arg.Name)
	var i Character
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Name,
		&i.PositionX,
		&i.PositionY,
		&i.ActionID,
		&i.ActionTarget,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getCharacterById = `-- name: GetCharacterById :one
SELECT id, user_id, name, position_x, position_y, action_id, action_target, created_at, updated_at from characters
WHERE id = $1
`

func (q *Queries) GetCharacterById(ctx context.Context, id pgtype.UUID) (Character, error) {
	row := q.db.QueryRow(ctx, getCharacterById, id)
	var i Character
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Name,
		&i.PositionX,
		&i.PositionY,
		&i.ActionID,
		&i.ActionTarget,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getCharacterByName = `-- name: GetCharacterByName :one
SELECT id, user_id, name, position_x, position_y, action_id, action_target, created_at, updated_at from characters
where name = $1
`

func (q *Queries) GetCharacterByName(ctx context.Context, name string) (Character, error) {
	row := q.db.QueryRow(ctx, getCharacterByName, name)
	var i Character
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Name,
		&i.PositionX,
		&i.PositionY,
		&i.ActionID,
		&i.ActionTarget,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateCharacterById = `-- name: UpdateCharacterById :one
UPDATE characters
SET action_id = $1, 
	updated_at = NOW()
WHERE id = $2
RETURNING id, user_id, name, position_x, position_y, action_id, action_target, created_at, updated_at
`

type UpdateCharacterByIdParams struct {
	ActionID int32
	ID       pgtype.UUID
}

func (q *Queries) UpdateCharacterById(ctx context.Context, arg UpdateCharacterByIdParams) (Character, error) {
	row := q.db.QueryRow(ctx, updateCharacterById, arg.ActionID, arg.ID)
	var i Character
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Name,
		&i.PositionX,
		&i.PositionY,
		&i.ActionID,
		&i.ActionTarget,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}