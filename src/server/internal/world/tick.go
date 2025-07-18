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

type TickUpdate struct {
	InventoryUpdate *api.InventoryUpdate
	ProgressUpdate  *api.CharacterProgressUpdate
}

type WorldConfig struct {
	DB       *database.Queries
	Redis    *redis.Client
	TickRate time.Duration
	Seed     *rand.Rand
	*api.ApiConfig
}

func (cfg *WorldConfig) ProcessTicks() {
	ticker := time.NewTicker(cfg.TickRate)
	defer ticker.Stop()

	for range ticker.C {
		activeChars, err := cfg.GetActiveCharacters(context.Background())
		if err != nil {
			log.Printf("Error getting active characters: %v", err)
			continue
		}

		updateChan := make(chan TickUpdate, len(activeChars))
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
		var progressUpdates []api.CharacterProgressUpdate
		for update := range updateChan {
			if update.InventoryUpdate != nil {
				inventoryUpdates = append(inventoryUpdates, *update.InventoryUpdate)
			}
			if update.ProgressUpdate != nil {
				progressUpdates = append(progressUpdates, *update.ProgressUpdate)
			}
		}

		if len(inventoryUpdates) > 0 {
			err := cfg.BatchAddItemsToInventory(context.Background(), inventoryUpdates)
			if err != nil {
				log.Printf("Error batch updating inventories: %v", err)
			}
		}

		if len(progressUpdates) > 0 {
			err := cfg.ApiConfig.BatchUpdateCharacterProgress(context.Background(), progressUpdates)
			if err != nil {
				log.Printf("Error batch updating character progress: %v", err)
			}
		}
	}
}

func (cfg *WorldConfig) processCharacterAction(char database.Character) *TickUpdate {
	return cfg.processResourceGathering(char)
}

func (cfg *WorldConfig) processResourceGathering(char database.Character) *TickUpdate {
	ctx := context.Background()

	if !char.ActionTarget.Valid {
		log.Printf("Character %s has no action target for resource gathering", char.Name)
		return nil
	}

	// Check if character has reached their gathering limit
	if char.ActionAmountLimit.Valid && char.ActionAmountProgress.Valid {
		if char.ActionAmountProgress.Int32 >= char.ActionAmountLimit.Int32 {
			err := cfg.ApiConfig.SetCharacterToIdle(ctx, char.ID)
			if err != nil {
				log.Printf("Failed to set character %s to idle: %v", char.Name, err)
			}
			message := fmt.Sprintf("Character %s finished gathering %d items and is now idle",
				char.Name, char.ActionAmountLimit.Int32)
			cfg.ApiConfig.Hub.SendNotificationToUser(char.UserID.Bytes, message, "info")
			return nil
		}
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

	result := &TickUpdate{
		InventoryUpdate: &api.InventoryUpdate{
			InventoryID: inventory.ID,
			ItemID:      drop.ItemID,
			Quantity:    1,
		},
	}

	// Add progress update if character has a limit set
	if char.ActionAmountLimit.Valid {
		newProgress := char.ActionAmountProgress.Int32 + 1
		result.ProgressUpdate = &api.CharacterProgressUpdate{
			CharacterID: char.ID,
			Progress:    newProgress,
		}
	}

	return result
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
