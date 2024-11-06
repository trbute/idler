package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/trbute/idler/internal/auth"
	"github.com/trbute/idler/internal/database"
	"net/http"
	"time"
)

type Character struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	ActionID  int32     `json:"action_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (cfg *apiConfig) handleCreateCharacter(w http.ResponseWriter, r *http.Request) {
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

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token invalid", err)
		return
	}

	fmt.Printf("%v \n", userID)

	character, err := cfg.db.CreateCharacter(r.Context(), database.CreateCharacterParams{
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

func (cfg *apiConfig) handleUpdateCharacter(w http.ResponseWriter, r *http.Request) {
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

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token invalid", err)
		return
	}

	character, err := cfg.db.GetCharacterByName(r.Context(), params.Name)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve user", err)
		return
	}

	if character.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Character doesn't belong to user", err)
		return
	}

	action, err := cfg.db.GetActionByID(r.Context(), params.ActionID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve action", err)
		return
	}

	character, err = cfg.db.UpdateCharacterByID(r.Context(), database.UpdateCharacterByIDParams{
		ID:       character.ID,
		ActionID: action.ID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database update failed", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Character{
		ID:        character.ID,
		ActionID:  character.ActionID,
		CreatedAt: character.CreatedAt,
		UpdatedAt: character.UpdatedAt,
	})
}
