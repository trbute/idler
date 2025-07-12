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

type Character struct {
	ID        pgtype.UUID      `json:"id"`
	UserID    pgtype.UUID      `json:"user_id"`
	Name      string           `json:"name"`
	ActionID  int32            `json:"action_id"`
	CreatedAt pgtype.Timestamp `json:"created_at"`
	UpdatedAt pgtype.Timestamp `json:"updated_at"`
}

func (cfg *ApiConfig) handleCreateCharacter(w http.ResponseWriter, r *http.Request) {
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

	type parameters struct {
		Name string `json:"name"`
	}
	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to decode parameters", err)
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

	_, err = cfg.DB.CreateInventory(r.Context(), database.CreateInventoryParams{
		CharacterID: character.ID,
		PositionX:   character.PositionX,
		PositionY:   character.PositionY,
		Capacity:    50,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Inventory creation failed", err)
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

	type parameters struct {
		CharacterName string `json:"character_name"`
		Target        string `json:"target"`
	}
	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		log.Printf("JSON decode error: %v", err)
		respondWithError(w, http.StatusBadRequest, "Unable to decode parameters", err)
		return
	}

	
	if params.CharacterName == "" {
		respondWithError(w, http.StatusBadRequest, "Character name is required", nil)
		return
	}
	
	character, err := cfg.GetCharacterByName(r.Context(), params.CharacterName)
	if err != nil {
		log.Printf("Character lookup failed for '%s': %v", params.CharacterName, err)
		respondWithError(w, http.StatusInternalServerError, "Unable to find character", err)
		return
	}

	// Find target resource node and get its action
	var action database.Action
	var actionTarget pgtype.Int4
	var foundNode *database.ResourceNode
	
	if params.Target != "" {
		resourceNodes, err := cfg.GetResourceNodeSpawnsByCoordinates(r.Context(), character.PositionX, character.PositionY)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Unable to get resource nodes", err)
			return
		}

		var foundSpawn *database.ResourceNodeSpawn
		for _, spawn := range resourceNodes {
			node, err := cfg.GetResourceNodeById(r.Context(), spawn.NodeID)
			if err == nil && strings.EqualFold(node.Name, params.Target) {
				foundNode = &node
				foundSpawn = &spawn
				break
			}
		}

		if foundNode == nil {
			respondWithError(w, http.StatusBadRequest, "Target not found at character location", nil)
			return
		}

		action, err = cfg.GetActionById(r.Context(), foundNode.ActionID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Unable to get action for target", err)
			return
		}

		actionTarget = pgtype.Int4{Int32: foundSpawn.ID, Valid: true}
	} else {
		respondWithError(w, http.StatusBadRequest, "Target must be provided", nil)
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

	// Check if action requires a tool and if character has one with sufficient tier
	if action.RequiredToolTypeID.Valid {
		bestTool, bestTier, err := cfg.GetBestToolForType(r.Context(), character.ID, action.RequiredToolTypeID.Int32, foundNode.MinToolTier)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Unable to check required tool", err)
			return
		}
		if bestTool == nil {
			toolType, err := cfg.DB.GetToolTypeById(r.Context(), action.RequiredToolTypeID.Int32)
			if err != nil {
				respondWithError(w, http.StatusBadRequest, "You need a required tool to perform this action", nil)
			} else {
				respondWithError(w, http.StatusBadRequest, fmt.Sprintf("You need a tier %d+ %s to perform this action", foundNode.MinToolTier, toolType.Name), nil)
			}
			return
		}
		
		// Store the tool tier for potential use in resource calculation
		_ = bestTier // We'll use this later for multipliers
	}


	char, err := cfg.DB.UpdateCharacterByIdWithTarget(r.Context(), database.UpdateCharacterByIdWithTargetParams{
		ActionID:     action.ID,
		ActionTarget: actionTarget,
		ID:           character.ID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Character update failed", err)
		return
	}


	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"name":        char.Name,
		"action_id":   char.ActionID,
		"action_name": action.Name,
		"target":      params.Target,
	})
}

func (cfg *ApiConfig) GetActiveCharacters(ctx context.Context) ([]database.Character, error) {
	return cfg.DB.GetActiveCharacters(ctx)
}

func (cfg *ApiConfig) GetCharacterByName(ctx context.Context, name string) (database.Character, error) {
	cacheKey := fmt.Sprintf("character:name:%s", name)

	cached, err := cfg.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var character database.Character
		if json.Unmarshal([]byte(cached), &character) == nil {
			return character, nil
		}
	}

	character, err := cfg.DB.GetCharacterByName(ctx, name)
	if err != nil {
		return database.Character{}, err
	}

	if data, err := json.Marshal(character); err == nil {
		cfg.Redis.Set(ctx, cacheKey, data, 30*time.Second)
	}

	return character, nil
}

func (cfg *ApiConfig) GetCharacterById(ctx context.Context, id pgtype.UUID) (database.Character, error) {
	cacheKey := fmt.Sprintf("character:id:%s", id.String())

	cached, err := cfg.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var character database.Character
		if json.Unmarshal([]byte(cached), &character) == nil {
			return character, nil
		}
	}

	character, err := cfg.DB.GetCharacterById(ctx, id)
	if err != nil {
		return database.Character{}, err
	}

	if data, err := json.Marshal(character); err == nil {
		cfg.Redis.Set(ctx, cacheKey, data, 30*time.Second)
	}

	return character, nil
}
