package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trbute/idler/server/internal/auth"
	"github.com/trbute/idler/server/internal/database"
)

type inventoryResponse struct {
	Items map[string]int32 `json:"items"`
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
	char, err := cfg.GetCharacterByName(r.Context(), charName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve character", err)
		return
	}

	pgUserId := pgtype.UUID{
		Bytes: userId,
		Valid: true,
	}

	if char.UserID != pgUserId {
		respondWithError(w, http.StatusUnauthorized, "Character doesn't belong to user", err)
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
		Items: items,
	}

	respondWithJSON(w, http.StatusOK, res)
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

	inventoryIDs := make([]pgtype.UUID, len(updates))
	itemIDs := make([]int32, len(updates))
	quantities := make([]int32, len(updates))

	for i, update := range updates {
		inventoryIDs[i] = update.InventoryID
		itemIDs[i] = update.ItemID
		quantities[i] = update.Quantity
	}

	return cfg.DB.BatchAddItemsToInventory(ctx, database.BatchAddItemsToInventoryParams{
		Column1: inventoryIDs,
		Column2: itemIDs,
		Column3: quantities,
	})
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
