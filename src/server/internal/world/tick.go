package world

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/trbute/idler/server/internal/database"
)

func (cfg *WorldConfig) ProcessTicks() {
	fmt.Println("starting tick")
	tick := time.NewTicker(time.Millisecond * 500)
	for range tick.C {
		for coord := range cfg.World.Grid {
			go cfg.processActions(coord)
		}
	}
}

func (cfg *WorldConfig) processActions(coord Coord) {
	cell := cfg.World.Grid[coord]
	chars := cell.Characters
	for _, char := range chars {
		switch char.ActionId {
		case 0:
			continue
		case 1:
			go cfg.processWoodCutting(cell, char.Name, char.ActionTarget)
		}
	}
}

func (cfg *WorldConfig) processWoodCutting(cell Cell, charName string, nodeName string) {
	fmt.Printf(
		"%v at %v, %v is choppin wood at %v\n",
		charName,
		cell.PositionX,
		cell.PositionY,
		nodeName,
	)

	drop := cfg.rollDrop(cell.ResourceNodes[nodeName].Resources)
	charItems := cell.Characters[charName].Inventory.Items
	_, ok := charItems[drop.Name]
	if !ok {
		charItems[drop.Name] = &Item{
			Name:     drop.Name,
			Quantity: 1,
		}
	} else {
		charItems[drop.Name].Quantity++
	}

	_, err := cfg.DB.AddItemsToInventory(context.Background(), database.AddItemsToInventoryParams{
		InventoryID: cell.Characters[charName].Inventory.InventoryID,
		ItemID:      drop.ID,
		Quantity:    charItems[drop.Name].Quantity,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func (cfg *WorldConfig) rollDrop(resources []Resource) Item {
	totalChance := 0
	for _, resource := range resources {
		totalChance += int(resource.DropChance)
	}

	n := rand.Intn(totalChance)

	for _, resource := range resources {
		n -= int(resource.DropChance)
		if n < 0 {
			return resource.Item
		}
	}

	return resources[0].Item
}
