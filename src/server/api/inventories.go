package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trbute/idler/server/internal/auth"
	"github.com/trbute/idler/server/internal/database"
)

type inventoryResponse struct {
	Items    map[string]int32 `json:"items"`
	Weight   int32            `json:"weight"`
	Capacity int32            `json:"capacity"`
}

func (cfg *ApiConfig) handleGetInventory(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to retrieve token", err)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.JwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token invalid", err)
		return
	}

	charName := r.PathValue("character")
	
	char, err := cfg.GetCharacterWithOwnershipValidation(r.Context(), charName, userId)
	if err != nil {
		if err.Error() == "character doesn't belong to user" {
			respondWithError(w, http.StatusUnauthorized, "Character doesn't belong to user", nil)
		} else {
			respondWithError(w, http.StatusInternalServerError, "Unable to retrieve character", err)
		}
		return
	}

	inventory, err := cfg.GetInventoryByCharacterId(r.Context(), char.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve inventory", err)
		return
	}

	inventoryItems, err := cfg.DB.GetInventoryItemsByInventoryId(r.Context(), inventory.ID)
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"Unable to retrieve inventory items",
			err,
		)
		return
	}

	items := map[string]int32{}
	for _, item := range inventoryItems {
		itemData, err := cfg.GetItemById(r.Context(), item.ItemID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Unable to retrieve item", err)
			return
		}
		items[itemData.Name] = item.Quantity
	}

	res := inventoryResponse{
		Items:    items,
		Weight:   inventory.Weight,
		Capacity: inventory.Capacity,
	}

	respondWithJSON(w, http.StatusOK, res)
}

func (cfg *ApiConfig) handleDropItem(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to retrieve token", err)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.JwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token invalid", err)
		return
	}

	type parameters struct {
		CharacterName string `json:"character_name"`
		ItemName      string `json:"item_name"`
		Quantity      int32  `json:"quantity"`
	}
	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to decode parameters", err)
		return
	}

	if params.CharacterName == "" {
		respondWithError(w, http.StatusBadRequest, "Character name is required", nil)
		return
	}

	if params.ItemName == "" {
		respondWithError(w, http.StatusBadRequest, "Item name is required", nil)
		return
	}

	if params.Quantity <= 0 {
		respondWithError(w, http.StatusBadRequest, "Quantity must be greater than 0", nil)
		return
	}

	char, err := cfg.GetCharacterWithOwnershipValidation(r.Context(), params.CharacterName, userId)
	if err != nil {
		if err.Error() == "character doesn't belong to user" {
			respondWithError(w, http.StatusUnauthorized, "Character doesn't belong to user", nil)
		} else {
			respondWithError(w, http.StatusInternalServerError, "Unable to retrieve character", err)
		}
		return
	}

	inventory, err := cfg.GetInventoryByCharacterId(r.Context(), char.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve inventory", err)
		return
	}

	item, err := cfg.GetItemByName(r.Context(), strings.ToUpper(params.ItemName))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Item not found", err)
		return
	}

	err = cfg.DropItemFromInventory(r.Context(), inventory.ID, item.ID, params.Quantity)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to drop item: "+err.Error(), err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": fmt.Sprintf("Dropped %d %s", params.Quantity, strings.Title(strings.ToLower(item.Name))),
	})
}

func (cfg *ApiConfig) GetInventoryByCharacterId(ctx context.Context, characterID pgtype.UUID) (database.Inventory, error) {
	cacheKey := fmt.Sprintf("inventory:char:%s", characterID.String())
	
	cached, err := cfg.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var inventory database.Inventory
		if json.Unmarshal([]byte(cached), &inventory) == nil {
			return inventory, nil
		}
	}
	
	inventory, err := cfg.DB.GetInventoryByCharacterId(ctx, characterID)
	if err != nil {
		return database.Inventory{}, err
	}
	
	if data, err := json.Marshal(inventory); err == nil {
		cfg.Redis.Set(ctx, cacheKey, data, 10*time.Second)
	}
	
	return inventory, nil
}

type InventoryUpdate struct {
	InventoryID pgtype.UUID
	ItemID      int32
	Quantity    int32
}

func (cfg *ApiConfig) BatchAddItemsToInventory(ctx context.Context, updates []InventoryUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	var validUpdates []InventoryUpdate
	for _, update := range updates {
		canAdd, err := cfg.CheckInventoryCapacity(ctx, update.InventoryID, update.ItemID, update.Quantity)
		if err != nil {
			return err
		}
		if !canAdd {
			cfg.SendInventoryFullNotification(ctx, update.InventoryID)
			continue
		}
		validUpdates = append(validUpdates, update)
	}

	if len(validUpdates) == 0 {
		return nil
	}

	inventoryIDs := make([]pgtype.UUID, len(validUpdates))
	itemIDs := make([]int32, len(validUpdates))
	quantities := make([]int32, len(validUpdates))

	for i, update := range validUpdates {
		inventoryIDs[i] = update.InventoryID
		itemIDs[i] = update.ItemID
		quantities[i] = update.Quantity
	}

	err := cfg.DB.BatchAddItemsToInventory(ctx, database.BatchAddItemsToInventoryParams{
		Column1: inventoryIDs,
		Column2: itemIDs,
		Column3: quantities,
	})
	if err != nil {
		return err
	}

	inventoryWeightUpdates := make(map[pgtype.UUID]int32)
	for _, update := range validUpdates {
		item, err := cfg.GetItemById(ctx, update.ItemID)
		if err != nil {
			continue
		}
		weightToAdd := item.Weight * update.Quantity
		inventoryWeightUpdates[update.InventoryID] += weightToAdd
	}

	for inventoryID, weightToAdd := range inventoryWeightUpdates {
		err := cfg.UpdateInventoryWeight(ctx, inventoryID, weightToAdd)
		if err != nil {
			log.Printf("Error updating inventory weight for %s: %v", inventoryID, err)
		}
	}

	return nil
}

func (cfg *ApiConfig) GetBestToolForType(ctx context.Context, characterID pgtype.UUID, toolTypeID int32, minTier int32) (*database.Item, int32, error) {
	inventory, err := cfg.GetInventoryByCharacterId(ctx, characterID)
	if err != nil {
		return nil, 0, err
	}

	inventoryItems, err := cfg.DB.GetInventoryItemsByInventoryId(ctx, inventory.ID)
	if err != nil {
		return nil, 0, err
	}

	var bestTool *database.Item
	var bestTier int32 = 0

	for _, invItem := range inventoryItems {
		if invItem.Quantity > 0 {
			item, err := cfg.GetItemById(ctx, invItem.ItemID)
			if err != nil {
				continue
			}

			if item.ToolTypeID.Valid && item.ToolTypeID.Int32 == toolTypeID {
				toolType, err := cfg.DB.GetToolTypeById(ctx, item.ToolTypeID.Int32)
				if err != nil {
					continue
				}

				if toolType.Tier >= minTier && toolType.Tier > bestTier {
					bestTool = &item
					bestTier = toolType.Tier
				}
			}
		}
	}

	if bestTool == nil {
		return nil, 0, nil
	}

	return bestTool, bestTier, nil
}

func (cfg *ApiConfig) CheckInventoryCapacity(ctx context.Context, inventoryID pgtype.UUID, itemID int32, quantity int32) (bool, error) {
	inventory, err := cfg.DB.GetInventory(ctx, inventoryID)
	if err != nil {
		return false, err
	}

	item, err := cfg.GetItemById(ctx, itemID)
	if err != nil {
		return false, err
	}

	totalWeightToAdd := item.Weight * quantity
	newWeight := inventory.Weight + totalWeightToAdd

	return newWeight <= inventory.Capacity, nil
}

func (cfg *ApiConfig) UpdateInventoryWeight(ctx context.Context, inventoryID pgtype.UUID, weightToAdd int32) error {
	err := cfg.DB.UpdateInventoryWeight(ctx, database.UpdateInventoryWeightParams{
		ID:     inventoryID,
		Weight: weightToAdd,
	})
	if err != nil {
		return err
	}

	inventory, err := cfg.DB.GetInventory(ctx, inventoryID)
	if err != nil {
		return err
	}

	if inventory.CharacterID.Valid {
		cacheKey := fmt.Sprintf("inventory:char:%s", inventory.CharacterID.String())
		cfg.Redis.Del(ctx, cacheKey)
	}

	return nil
}

func (cfg *ApiConfig) DropItemFromInventory(ctx context.Context, inventoryID pgtype.UUID, itemID int32, quantity int32) error {
	err := cfg.DB.RemoveItemsFromInventory(ctx, database.RemoveItemsFromInventoryParams{
		InventoryID: inventoryID,
		ItemID:      itemID,
		Quantity:    quantity,
	})
	if err != nil {
		return fmt.Errorf("insufficient quantity or item not found")
	}

	err = cfg.DB.DeleteEmptyInventoryItems(ctx, inventoryID)
	if err != nil {
		log.Printf("Error cleaning up empty inventory items: %v", err)
	}

	item, err := cfg.GetItemById(ctx, itemID)
	if err != nil {
		return err
	}

	weightToRemove := -(item.Weight * quantity)
	err = cfg.UpdateInventoryWeight(ctx, inventoryID, weightToRemove)
	if err != nil {
		log.Printf("Error updating inventory weight after drop: %v", err)
	}

	return nil
}

func (cfg *ApiConfig) SendInventoryFullNotification(ctx context.Context, inventoryID pgtype.UUID) {
	inventory, err := cfg.DB.GetInventory(ctx, inventoryID)
	if err != nil {
		return
	}

	if !inventory.CharacterID.Valid {
		return
	}

	character, err := cfg.GetCharacterById(ctx, inventory.CharacterID)
	if err != nil {
		return
	}

	err = cfg.SetCharacterToIdle(ctx, character.ID)
	if err != nil {
		log.Printf("Failed to set character %s to idle: %v", character.Name, err)
	}

	message := fmt.Sprintf("Inventory is full for character %s! Character set to idle.", character.Name)
	cfg.Hub.SendNotificationToUser(character.UserID.Bytes, message, "warning")
}
