// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package database

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Action struct {
	ID   int32
	Name string
}

type Character struct {
	ID           pgtype.UUID
	UserID       pgtype.UUID
	Name         string
	PositionX    int32
	PositionY    int32
	ActionID     int32
	ActionTarget pgtype.UUID
	CreatedAt    pgtype.Timestamp
	UpdatedAt    pgtype.Timestamp
}

type Grid struct {
	PositionX int32
	PositionY int32
}

type Inventory struct {
	ID          pgtype.UUID
	CharacterID pgtype.UUID
	PositionX   int32
	PositionY   int32
	Weight      int32
	Capacity    int32
	CreatedAt   pgtype.Timestamp
	UpdatedAt   pgtype.Timestamp
}

type InventoryItem struct {
	ID          pgtype.UUID
	ItemID      int32
	InventoryID pgtype.UUID
	Quantity    int32
	CreatedAt   pgtype.Timestamp
	UpdatedAt   pgtype.Timestamp
}

type Item struct {
	ID     int32
	Name   string
	Weight int32
}

type RefreshToken struct {
	Token     string
	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
	UserID    pgtype.UUID
	ExpiresAt pgtype.Timestamp
	RevokedAt pgtype.Timestamp
}

type Resource struct {
	ID             int32
	ResourceNodeID int32
	ItemID         int32
	DropChance     int32
}

type ResourceNode struct {
	ID       int32
	Name     pgtype.Text
	ActionID int32
	Tier     int32
}

type ResourceNodeSpawn struct {
	ID        int32
	NodeID    int32
	PositionX int32
	PositionY int32
}

type User struct {
	ID             pgtype.UUID
	Email          string
	HashedPassword string
	CreatedAt      pgtype.Timestamp
	UpdatedAt      pgtype.Timestamp
}
