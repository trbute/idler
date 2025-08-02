package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trbute/idler/server/internal/auth"
	"github.com/trbute/idler/server/internal/database"
	"github.com/trbute/idler/server/internal/validation"
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

	userID, err := auth.ValidateJWTWithBlacklist(r.Context(), token, cfg.JwtSecret, cfg.Redis)
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

	if err := validation.ValidateCharacterName(params.Name); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), nil)
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

	userID, err := auth.ValidateJWTWithBlacklist(r.Context(), token, cfg.JwtSecret, cfg.Redis)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token invalid", err)
		return
	}

	type parameters struct {
		CharacterName string `json:"character_name"`
		Target        string `json:"target"`
		Amount        *int   `json:"amount,omitempty"`
	}
	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to decode parameters", err)
		return
	}

	if err := validation.ValidateCharacterName(params.CharacterName); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	params.Target = strings.TrimSpace(strings.ToUpper(params.Target))
	if err := validation.ValidateTarget(params.Target); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := validation.ValidateAmount(params.Amount); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	character, err := cfg.GetCharacterWithOwnershipValidation(r.Context(), params.CharacterName, userID)
	if err != nil {
		if err.Error() == "character doesn't belong to user" {
			respondWithError(w, http.StatusUnauthorized, "Character doesn't belong to user", nil)
		} else {
			respondWithError(w, http.StatusInternalServerError, "Unable to find character", err)
		}
		return
	}

	var action database.Action
	var actionTarget pgtype.Int4
	var foundNode *database.ResourceNode

	if params.Target == "IDLE" {
		action, err = cfg.GetActionByName(r.Context(), "IDLE")
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Unable to get idle action", err)
			return
		}
		actionTarget = pgtype.Int4{Valid: false}
	} else if params.Target != "" {
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

	if action.RequiredToolTypeID.Valid && params.Target != "IDLE" {
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

		_ = bestTier
	}

	var amountLimit pgtype.Int4
	if params.Amount != nil && *params.Amount > 0 {
		amountLimit = pgtype.Int4{Int32: int32(*params.Amount), Valid: true}
	} else {
		amountLimit = pgtype.Int4{Valid: false} // NULL - resets any existing limit
	}

	char, err := cfg.DB.UpdateCharacterByIdWithTargetAndAmount(r.Context(), database.UpdateCharacterByIdWithTargetAndAmountParams{
		ActionID:          action.ID,
		ActionTarget:      actionTarget,
		ActionAmountLimit: amountLimit,
		ID:                character.ID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Character update failed", err)
		return
	}

	// Invalidate active characters cache since character action changed
	cfg.InvalidateActiveCharactersCache(r.Context())

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"name":        char.Name,
		"action_id":   char.ActionID,
		"action_name": action.Name,
		"target":      params.Target,
	})
}

func (cfg *ApiConfig) GetActiveCharacters(ctx context.Context) ([]database.Character, error) {
	cacheKey := "active_characters"
	
	// Try to get from cache first
	cached, err := cfg.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var characters []database.Character
		if json.Unmarshal([]byte(cached), &characters) == nil {
			return characters, nil
		}
	}
	
	// Cache miss - get from database
	characters, err := cfg.DB.GetActiveCharacters(ctx)
	if err != nil {
		return nil, err
	}
	
	// Cache the result for 30 seconds
	if data, err := json.Marshal(characters); err == nil {
		cfg.Redis.Set(ctx, cacheKey, data, 30*time.Second)
	}
	
	return characters, nil
}

func (cfg *ApiConfig) BatchUpdateCharacterProgress(ctx context.Context, updates []CharacterProgressUpdate) error {
	if len(updates) == 0 {
		return nil
	}
	
	ids := make([]pgtype.UUID, len(updates))
	progress := make([]int32, len(updates))
	
	for i, update := range updates {
		ids[i] = update.CharacterID
		progress[i] = update.Progress
	}
	
	return cfg.DB.BatchUpdateCharacterProgress(ctx, database.BatchUpdateCharacterProgressParams{
		Column1: ids,
		Column2: progress,
	})
}

type CharacterProgressUpdate struct {
	CharacterID pgtype.UUID
	Progress    int32
}

func (cfg *ApiConfig) InvalidateActiveCharactersCache(ctx context.Context) {
	cfg.Redis.Del(ctx, "active_characters")
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

func (cfg *ApiConfig) ValidateCharacterOwnership(ctx context.Context, characterName string, userID uuid.UUID) (bool, error) {
	character, err := cfg.GetCharacterByName(ctx, characterName)
	if err != nil {
		return false, err
	}

	pgUserID := pgtype.UUID{
		Bytes: userID,
		Valid: true,
	}

	return character.UserID == pgUserID, nil
}

func (cfg *ApiConfig) GetCharacterWithOwnershipValidation(ctx context.Context, characterName string, userID uuid.UUID) (database.Character, error) {
	character, err := cfg.GetCharacterByName(ctx, characterName)
	if err != nil {
		return database.Character{}, err
	}

	pgUserID := pgtype.UUID{
		Bytes: userID,
		Valid: true,
	}

	if character.UserID != pgUserID {
		return database.Character{}, fmt.Errorf("character doesn't belong to user")
	}

	return character, nil
}

func (cfg *ApiConfig) SetCharacterToIdle(ctx context.Context, characterID pgtype.UUID) error {
	idleAction, err := cfg.GetActionByName(ctx, "IDLE")
	if err != nil {
		return err
	}

	_, err = cfg.DB.SetCharacterToIdleAndResetGathering(ctx, database.SetCharacterToIdleAndResetGatheringParams{
		ActionID: idleAction.ID,
		ID:       characterID,
	})
	if err == nil {
		// Invalidate active characters cache since character went idle
		cfg.InvalidateActiveCharactersCache(ctx)
	}
	return err
}

func (cfg *ApiConfig) handleSelectCharacter(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to retrieve token", err)
		return
	}

	userID, err := auth.ValidateJWTWithBlacklist(r.Context(), token, cfg.JwtSecret, cfg.Redis)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token invalid", err)
		return
	}

	characterName := r.PathValue("character")
	if err := validation.ValidateCharacterName(characterName); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	character, err := cfg.GetCharacterWithOwnershipValidation(r.Context(), characterName, userID)
	if err != nil {
		if err.Error() == "character doesn't belong to user" {
			respondWithError(w, http.StatusUnauthorized, "Character doesn't belong to user", nil)
		} else {
			respondWithError(w, http.StatusNotFound, "Character not found", err)
		}
		return
	}

	respondWithJSON(w, http.StatusOK, Character{
		ID:        character.ID,
		UserID:    character.UserID,
		Name:      character.Name,
		ActionID:  character.ActionID,
		CreatedAt: character.CreatedAt,
		UpdatedAt: character.UpdatedAt,
	})
}
