package world

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/trbute/idler/internal/database"
)

type WorldConfig struct {
	DB       *database.Queries
	Platform string
	TickRate time.Duration
	World    *World
}

type Item struct {
	ID   uuid.UUID
	Name string
}

type Inventory struct {
	Items []Item
}

type Character struct {
	Name         string
	ActionId     int32
	Inventory    Inventory
	LastActionAt time.Time
}

type Resource struct {
	Item       Item
	DropChance float64
}

type ResourceNode struct {
	Name      string
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
	ResourceNodes []ResourceNode
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

	resourceNodeRecords, err := cfg.DB.GetResourceNodes(ctx)
	if err != nil {
		log.Fatalf("Failed to get resource nodes: %v", err)
	}

	for _, resourceNodeRecord := range resourceNodeRecords {
		resourceNodeItem := ResourceNode{}
		var resourceNodeName string
		if resourceNodeRecord.Name.Valid {
			resourceNodeName = resourceNodeRecord.Name.String
		} else {
			log.Fatalf("Resource node record name is not a valid string")
		}

		resourceNodeItem.Name = resourceNodeName

		key := Coord{
			resourceNodeRecord.PositionX,
			resourceNodeRecord.PositionY,
		}

		resourceRecords, err := cfg.DB.GetResourcesByNodeId(ctx, resourceNodeRecord.ID)
		if err != nil {
			log.Fatalf("Failed to get resources: %v", err)
		}

		for _, resourceRecord := range resourceRecords {
			resourceItem := Resource{}
			resourceItem.DropChance, err = strconv.ParseFloat(resourceRecord.DropChance, 64)
			if err != nil {
				log.Fatalf("Failed to parse drop chance as float: %v", err)
			}

			itemItem := Item{}
			itemRecord, err := cfg.DB.GetItemByResourceId(ctx, resourceRecord.ID)
			if err != nil {
				log.Fatalf("Failed to get resources: %v", err)
			}

			itemItem.ID = itemRecord.ID
			itemItem.Name = itemRecord.Name
			resourceItem.Item = itemItem

			resourceNodeItem.Resources = append(resourceNodeItem.Resources, resourceItem)

		}

		gridCell := world.Grid[key]
		gridCell.ResourceNodes = append(gridCell.ResourceNodes, resourceNodeItem)
		world.Grid[key] = gridCell
	}

	duration := time.Since(start)
	fmt.Printf("Finished worldbuilding in %s\n", duration)

	return world
}
