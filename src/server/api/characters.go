package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trbute/idler/server/internal/auth"
	"github.com/trbute/idler/server/internal/database"
	"github.com/trbute/idler/server/internal/world"
)

type Character struct {
	ID        pgtype.UUID      `json:"id"`
	UserID    pgtype.UUID      `json:"user_id"`
	Name      string           `json:"name"`
	ActionID  int32            `json:"action_id"`
	CreatedAt pgtype.Timestamp `json:"created_at"`
	UpdatedAt pgtype.Timestamp `json:"updated_at"`
}

func (cfg *ApiConfig) handleCreateCharacter(w http.ResponseWriter, r *http.Request) {
	type Parameters struct {
		Name string `json:"name"`
	}

	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to decode parameters", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to retrieve token", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.JwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token invalid", err)
		return
	}

	pgUserID := pgtype.UUID{
		Bytes: userID,
		Valid: true,
	}

	character, err := cfg.DB.CreateCharacter(r.Context(), database.CreateCharacterParams{
		UserID: pgUserID,
		Name:   params.Name,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Character creation failed", err)
		return
	}

	_, err = cfg.DB.CreateInventory(
		r.Context(),
		database.CreateInventoryParams{
			CharacterID: character.ID,
			PositionX:   character.PositionX,
			PositionY:   character.PositionY,
		},
	)
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"Character inventory creation failed",
			err,
		)
		return
	}

	respondWithJSON(w, http.StatusCreated, Character{
		ID:        character.ID,
		UserID:    character.UserID,
		Name:      character.Name,
		ActionID:  character.ActionID,
		CreatedAt: character.CreatedAt,
		UpdatedAt: character.UpdatedAt,
	})
}

func (cfg *ApiConfig) handleUpdateCharacter(w http.ResponseWriter, r *http.Request) {
	type Parameters struct {
		CharacterName string `json:"character_name"`
		ActionName    string `json:"action_name"`
		Target        string `json:"target"`
	}

	decoder := json.NewDecoder(r.Body)
	params := Parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to decode parameters", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to retrieve token", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.JwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token invalid", err)
		return
	}

	character, err := cfg.DB.GetCharacterByName(r.Context(), params.CharacterName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve character", err)
		return
	}

	resourceNodes := cfg.World.Grid[world.Coord{PositionX: character.PositionX, PositionY: character.PositionY}].ResourceNodes
	var ok bool
	node, ok := resourceNodes[params.Target]
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Unable to find target node", err)
		return
	}

	action, err := cfg.DB.GetActionByName(r.Context(), params.ActionName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve action", err)
		return
	}

	if node.ActionID != action.ID {
		respondWithError(w, http.StatusBadRequest, "Can't do that to this node type", err)
		return
	}

	pgUserID := pgtype.UUID{
		Bytes: userID,
		Valid: true,
	}

	if character.UserID != pgUserID {
		respondWithError(w, http.StatusUnauthorized, "Character doesn't belong to user", err)
		return
	}

	playerCoords := world.Coord{
		PositionX: character.PositionX,
		PositionY: character.PositionY,
	}

	chars := cfg.World.Grid[playerCoords].Characters

	var char *world.Character
	char, ok = chars[character.Name]
	if !ok {
		inventoryRecord, err := cfg.DB.GetInventoryByUserId(context.Background(), character.ID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Unable to retrieve inventory", err)
			return
		}

		inventoryItems, err := cfg.DB.GetInventoryItemsByInventoryId(
			context.Background(),
			inventoryRecord.ID,
		)
		if err != nil {
			respondWithError(
				w,
				http.StatusInternalServerError,
				"Unable to retrieve inventory items",
				err,
			)
			return
		}

		charInventory := world.Inventory{Items: make(map[string]*world.Item)}
		charInventory.InventoryID = inventoryRecord.ID
		for _, inventoryItem := range inventoryItems {
			item, err := cfg.DB.GetItemById(context.Background(), inventoryItem.ItemID)
			if err != nil {
				respondWithError(
					w,
					http.StatusInternalServerError,
					"Unable to retrieve inventory items",
					err,
				)
				return
			}

			charInventory.Items[item.Name] = &world.Item{
				ID:       item.ID,
				Name:     item.Name,
				Quantity: inventoryItem.Quantity,
			}
		}

		char = &world.Character{
			Name:         character.Name,
			ActionId:     action.ID,
			ActionTarget: params.Target,
			Inventory:    charInventory,
			LastActionAt: time.Now(),
		}
		chars[character.Name] = char
	}

	char.ActionId = action.ID
	char.LastActionAt = time.Now()

	respondWithJSON(w, http.StatusCreated, Character{
		Name:     char.Name,
		ActionID: char.ActionId,
	})
}
