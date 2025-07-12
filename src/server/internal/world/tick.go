package world

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/trbute/idler/server/api"
	"github.com/trbute/idler/server/internal/database"
)

type WorldConfig struct {
	DB       *database.Queries
	Redis    *redis.Client
	TickRate time.Duration
	Seed     *rand.Rand
	*api.ApiConfig
}

func (cfg *WorldConfig) ProcessTicks() {
	fmt.Println("starting tick system")
	ticker := time.NewTicker(cfg.TickRate)
	defer ticker.Stop()

	for range ticker.C {
		activeChars, err := cfg.GetActiveCharacters(context.Background())
		if err != nil {
			log.Printf("Error getting active characters: %v", err)
			continue
		}

		updateChan := make(chan api.InventoryUpdate, len(activeChars))
		var wg sync.WaitGroup

		for _, char := range activeChars {
			wg.Add(1)
			go func(char database.Character) {
				defer wg.Done()
				if update := cfg.processCharacterAction(char); update != nil {
					updateChan <- *update
				}
			}(char)
		}

		go func() {
			wg.Wait()
			close(updateChan)
		}()

		var inventoryUpdates []api.InventoryUpdate
		for update := range updateChan {
			inventoryUpdates = append(inventoryUpdates, update)
		}

		if len(inventoryUpdates) > 0 {
			err := cfg.BatchAddItemsToInventory(context.Background(), inventoryUpdates)
			if err != nil {
				log.Printf("Error batch updating inventories: %v", err)
			}
		}
	}
}

func (cfg *WorldConfig) processCharacterAction(char database.Character) *api.InventoryUpdate {
	return cfg.processResourceGathering(char)
}

func (cfg *WorldConfig) processResourceGathering(char database.Character) *api.InventoryUpdate {
	ctx := context.Background()

	if !char.ActionTarget.Valid {
		log.Printf("Character %s has no action target for resource gathering", char.Name)
		return nil
	}

	inventory, err := cfg.GetInventoryByCharacterId(ctx, char.ID)
	if err != nil {
		log.Printf("Error getting inventory for character %s: %v", char.Name, err)
		return nil
	}

	spawn, err := cfg.GetResourceNodeSpawnById(ctx, char.ActionTarget.Int32)
	if err != nil {
		log.Printf("Error getting target resource node spawn for character %s: %v", char.Name, err)
		return nil
	}

	resources, err := cfg.GetResourcesByNodeId(ctx, spawn.NodeID)
	if err != nil {
		log.Printf("Error getting resources for character %s: %v", char.Name, err)
		return nil
	}

	if len(resources) == 0 {
		log.Printf("No resources found for node %d for character %s", spawn.NodeID, char.Name)
		return nil
	}

	drop := cfg.rollDrop(resources)

	return &api.InventoryUpdate{
		InventoryID: inventory.ID,
		ItemID:      drop.ItemID,
		Quantity:    1,
	}
}

func (cfg *WorldConfig) rollDrop(resources []database.Resource) database.Resource {
	if len(resources) == 0 {
		return database.Resource{}
	}

	totalChance := 0
	for _, resource := range resources {
		totalChance += int(resource.DropChance)
	}

	if totalChance == 0 {
		return resources[0]
	}

	n := cfg.Seed.Intn(totalChance)

	for _, resource := range resources {
		n -= int(resource.DropChance)
		if n < 0 {
			return resource
		}
	}

	return resources[0]
}
