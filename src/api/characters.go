package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/trbute/idler/internal/auth"
	"github.com/trbute/idler/internal/database"
	"github.com/trbute/idler/internal/world"
)

type Character struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	ActionID  int32     `json:"action_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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

	character, err := cfg.DB.CreateCharacter(r.Context(), database.CreateCharacterParams{
		UserID: userID,
		Name:   params.Name,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Character creation failed", err)
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
		Name     string `json:"name"`
		ActionID int32  `json:"action_id"`
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

	character, err := cfg.DB.GetCharacterByName(r.Context(), params.Name)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve character", err)
		return
	}

	if character.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Character doesn't belong to user", err)
		return
	}

	action, err := cfg.DB.GetActionByID(r.Context(), params.ActionID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve action", err)
		return
	}

	playerCoords := world.Coord{
		PositionX: character.PositionX,
		PositionY: character.PositionY,
	}

	chars := cfg.World.Grid[playerCoords].Characters

	var char *world.Character
	var ok bool
	char, ok = chars[character.Name]
	if ok {
		char.ActionId = action.ID
		char.LastActionAt = time.Now()
	} else {
		char = &world.Character{
			Name:         character.Name,
			ActionId:     action.ID,
			Inventory:    world.Inventory{},
			LastActionAt: time.Now(),
		}
		chars[character.Name] = char
	}

	respondWithJSON(w, http.StatusOK, Character{
		Name:     char.Name,
		ActionID: char.ActionId,
	})
}
