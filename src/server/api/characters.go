package api

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5"
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

	// Start transaction using Pool
	txp, err := cfg.Pool.BeginTx(r.Context(), pgx.TxOptions{})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to start transaction", err)
		return
	}
	defer func() {
		if err != nil {
			txp.Rollback(r.Context())
		} else {
			txp.Commit(r.Context())
		}
	}()

	tx := cfg.DB.WithTx(txp)

	pgUserID := pgtype.UUID{
		Bytes: userID,
		Valid: true,
	}

	character, err := tx.CreateCharacter(r.Context(), database.CreateCharacterParams{
		UserID: pgUserID,
		Name:   params.Name,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Character creation failed", err)
		return
	}

	_, err = tx.CreateInventory(
		r.Context(),
		database.CreateInventoryParams{
			CharacterID: character.ID,
			PositionX:   character.PositionX,
			PositionY:   character.PositionY,
			Capacity:    64,
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

	txp, err := cfg.Pool.BeginTx(r.Context(), pgx.TxOptions{})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to start transaction", err)
		return
	}
	defer func() {
		if err != nil {
			txp.Rollback(r.Context())
		} else {
			txp.Commit(r.Context())
		}
	}()

	tx := cfg.DB.WithTx(txp)

	character, err := tx.GetCharacterByName(r.Context(), params.CharacterName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve character", err)
		return
	}

	node, err := tx.GetResourceNodeByName(r.Context(), params.Target)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve resource node", err)
		return
	}

	_, err = tx.GetResourceNodeSpawnByCoordsAndNodeId(r.Context(), database.GetResourceNodeSpawnByCoordsAndNodeIdParams{
		PositionX: character.PositionX,
		PositionY: character.PositionY,
		NodeID:    node.ID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to retrieve resource node spawn", err)
		return
	}

	action, err := tx.GetActionByName(r.Context(), params.ActionName)
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

	char, err := tx.UpdateCharacterById(r.Context(), database.UpdateCharacterByIdParams{
		ActionID: action.ID,
		ID:       character.ID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Character update failed", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Character{
		Name:     char.Name,
		ActionID: char.ActionID,
	})
}
