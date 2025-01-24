package world

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trbute/idler/server/internal/database"
)

type WorldConfig struct {
	DB       *database.Queries
	Platform string
	TickRate time.Duration
	Seed     *rand.Rand
	World    *World
}

type Item struct {
	ID       int32
	Name     string
	Quantity int32
	Weight   int32
}

type Inventory struct {
	InventoryID pgtype.UUID
	Items       map[string]*Item
	Weight      int32
	Capacity    int32
}

type Character struct {
	Name         string
	ActionId     int32
	ActionTarget string
	Inventory    Inventory
	LastActionAt time.Time
}

type Resource struct {
	Item       Item
	DropChance int32
}

type ResourceNode struct {
	Name      string
	ActionID  int32
	Resources []Resource
}

type Coord struct {
	PositionX int32
	PositionY int32
}

type Cell struct {
	PositionX     int32
	PositionY     int32
	Characters    map[string]*Character
	ResourceNodes map[string]ResourceNode
}

type World struct {
	Grid map[Coord]Cell
}

func (cfg *WorldConfig) GetWorld() *World {
	start := time.Now()
	fmt.Println("started worldbuilding")

	world := new(World)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grid, err := cfg.DB.GetGrid(ctx)
	if err != nil {
		log.Fatalf("Failed to get grid:%v", err)
	}

	for _, gridItem := range grid {
		key := Coord{
			PositionX: gridItem.PositionX,
			PositionY: gridItem.PositionY,
		}

		world.Grid = make(map[Coord]Cell)
		world.Grid[key] = Cell{
			PositionX:  gridItem.PositionX,
			PositionY:  gridItem.PositionY,
			Characters: make(map[string]*Character),
		}
	}

	resourceNodeSpawnRecords, err := cfg.DB.GetResourceNodeSpawns(ctx)
	if err != nil {
		log.Fatalf("Failed to get resource nodes: %v", err)
	}

	for _, resourceNodeSpawnRecord := range resourceNodeSpawnRecords {
		resourceNodeItem := ResourceNode{}
		resourceNodeRecord, err := cfg.DB.GetResourceNodeById(ctx, resourceNodeSpawnRecord.ID)
		if err != nil {
			log.Fatalf("Failed to get resource nodes: %v", err)
		}
		resourceNodeItem.Name = resourceNodeRecord.Name.String
		resourceNodeItem.ActionID = resourceNodeRecord.ActionID

		key := Coord{
			resourceNodeSpawnRecord.PositionX,
			resourceNodeSpawnRecord.PositionY,
		}

		resourceRecords, err := cfg.DB.GetResourcesByNodeId(ctx, resourceNodeSpawnRecord.ID)
		if err != nil {
			log.Fatalf("Failed to get resources: %v", err)
		}

		for _, resourceRecord := range resourceRecords {
			resourceItem := Resource{}
			resourceItem.DropChance = resourceRecord.DropChance

			itemRecord, err := cfg.DB.GetItemByResourceId(ctx, resourceRecord.ID)
			if err != nil {
				log.Fatalf("Failed to get resources: %v", err)
			}

			itemItem := Item{}
			itemItem.ID = itemRecord.ID
			itemItem.Name = itemRecord.Name
			itemItem.Weight = itemRecord.Weight
			resourceItem.Item = itemItem

			resourceNodeItem.Resources = append(resourceNodeItem.Resources, resourceItem)

		}

		gridCell := world.Grid[key]
		gridCell.ResourceNodes = make(map[string]ResourceNode)
		gridCell.ResourceNodes[resourceNodeItem.Name] = resourceNodeItem
		world.Grid[key] = gridCell
	}

	duration := time.Since(start)
	fmt.Printf("Finished worldbuilding in %s\n", duration)

	return world
}
