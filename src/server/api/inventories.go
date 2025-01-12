package api

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trbute/idler/server/internal/auth"
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
	char, err := cfg.DB.GetCharacterByName(r.Context(), charName)
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

	inventory, err := cfg.DB.GetInventoryByCharacterId(r.Context(), char.ID)
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
		itemData, err := cfg.DB.GetItemById(r.Context(), item.ItemID)
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
